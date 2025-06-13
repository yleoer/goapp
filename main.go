package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	redisAddr = "redis-service:6379"
	podID     = "HOSTNAME"
)

var taskHandlers = map[string]func(){
	"cleanup_report": func() {
		log.Printf("[%s] 执行报告清理任务", podID)
		time.Sleep(2 * time.Second)
	},
	"generate_stats": func() {
		log.Printf("[%s] 生产每日统计", podID)
	},
	"backup_data": func() {
		log.Printf("[%s] 数据备份任务", podID)
	},
}

// 生产示例任务
type task struct {
	id        string
	executeAt time.Time
}

func generateSampleTasks() []task {
	return []task{
		{
			id:        "cleanup_report",
			executeAt: time.Now().Add(70 * time.Second),
		},
		{
			id:        "generate_stats",
			executeAt: time.Now().Add(120 * time.Second),
		},
		{
			id:        "backup_data",
			executeAt: time.Now().Add(180 * time.Second),
		},
	}
}

// 任务执行器
func runTaskExecutor(rc *RedisClient) {
	ch := rc.SubscribeTaskEvent()

	for msg := range ch {
		taskID := msg.Payload

		if handler, exist := taskHandlers[taskID]; exist {
			log.Printf("[%s] 开始执行任务：%s", podID, taskID)
			handler()
			log.Printf("[%s] 任务完成：%s", podID, taskID)
		} else {
			log.Printf("[%s] 收到未知任务：%s", podID, taskID)
		}
	}
}

// 任务调度器（运行在多个Pod上）
func runScheduler(rc *RedisClient) {
	lockTTL := 5 * time.Second // 分布式锁有效期

	for {
		now := time.Now()
		nextMin := now.Truncate(time.Minute).Add(61 * time.Second)
		waitDuration := nextMin.Sub(now)

		time.Sleep(waitDuration)

		// 使用Redis SETNX实现分布式锁
		lockKey := "scheduler_lock"
		locked, err := rc.client.SetNX(
			context.Background(),
			lockKey,
			podID,
			lockTTL,
		).Result()

		if err == nil && locked {
			log.Printf("[%s] 获得调度器锁", podID)

			// 1. 生成定时任务（实际中可从数据库读取）
			tasks := generateSampleTasks()

			// 2. 添加新任务到Redis
			for _, task := range tasks {
				rc.AddTask(task.id, task.executeAt)
			}

			// 3. 处理到期任务
			dueTasks, _ := rc.GetDueTasks()
			for _, taskID := range dueTasks {
				rc.PublishTaskEvent(taskID)
				rc.RemoveTask(taskID)
			}

			log.Printf("[%s] 释放调度器锁", podID)
		}
	}
}

func main() {
	// 获取当前Pod的唯一标识（使用K8s HOSTNAME）
	if os.Getenv(podID) == "" {
		log.Fatal("HOSTNAME环境变量未设置")
	}

	// 初始化Redis客户端
	rc := NewRedisClient(redisAddr)
	defer rc.client.Close()

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			// 添加Redis连接检查
			if err := rc.client.Ping(context.Background()).Err(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("DOWN: Redis connection failed"))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("UP"))
		})

		// 添加metrics端点（可选）
		//http.Handle("/metrics", promhttp.Handler())

		log.Println("Health check server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start health server: %v", err)
		}
	}()

	// 启动任务调度器（每个Pod都运行）
	go runScheduler(rc)

	// 启动任务监听器（每个Pod都订阅）
	runTaskExecutor(rc)
}
