# 第一阶段：构建应用
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
ARG VERSION=1.0.0
ENV CGO_ENABLED=0

# 构建应用 (设置-ldflags参数减小二进制大小)
RUN go build -v -ldflags="-w -s -X main.BuildVersion=$VERSION" -o scheduler .

# 第二阶段：创建最小化运行时镜像
FROM alpine:3.18

# 安装必要的工具 (用于健康检查)
RUN apk add --no-cache curl

# 从构建阶段复制二进制文件
COPY --from=builder /app/scheduler /usr/local/bin/scheduler

# 设置健康检查
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露健康检查端口 (如果需要)
EXPOSE 8080

# 设置启动命令
CMD ["scheduler"]