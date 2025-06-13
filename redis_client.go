package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const scheduledTasks = "scheduled_tasks"

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{rdb}
}

// AddTask 添加定时任务到Sorted Set
func (rc *RedisClient) AddTask(taskID string, executeAt time.Time) error {
	z := redis.Z{
		Score:  float64(executeAt.Unix()),
		Member: taskID,
	}
	return rc.client.ZAdd(context.Background(), scheduledTasks, z).Err()
}

// GetDueTasks 获取到期的任务
func (rc *RedisClient) GetDueTasks() ([]string, error) {
	now := time.Now().Unix()
	return rc.client.ZRangeByScore(context.Background(), scheduledTasks, &redis.ZRangeBy{Min: "0", Max: fmt.Sprintf("%d", now)}).Result()
}

// RemoveTask 移除已执行任务
func (rc *RedisClient) RemoveTask(taskID string) error {
	return rc.client.ZRem(context.Background(), scheduledTasks, taskID).Err()
}

// PublishTaskEvent 发布任务触发事件
func (rc *RedisClient) PublishTaskEvent(taskID string) error {
	return rc.client.Publish(context.Background(), "task_events", taskID).Err()
}

func (rc *RedisClient) SubscribeTaskEvent() <-chan *redis.Message {
	pubsub := rc.client.Subscribe(context.Background(), "task_events")
	return pubsub.Channel()
}
