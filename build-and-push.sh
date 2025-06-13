#!/bin/bash

# 设置镜像仓库和名称
REPO_NAME="yleoer" # 修改为你的Docker Hub用户名
IMAGE_NAME="k8s-redis-scheduler"
TAG="1.0.0"

# 构建Docker镜像
echo "🛠️ 构建Docker镜像..."
docker build -t $REPO_NAME/$IMAGE_NAME:$TAG . --build-arg VERSION=$TAG

# 标记镜像为latest
echo "🏷️ 标记为latest..."
docker tag $REPO_NAME/$IMAGE_NAME:$TAG $REPO_NAME/$IMAGE_NAME:latest

# 登录Docker Hub
echo "🔑 登录Docker Hub..."
docker login -u $REPO_NAME

# 推送镜像
echo "🚀 推送镜像到仓库..."
docker push $REPO_NAME/$IMAGE_NAME:$TAG
docker push $REPO_NAME/$IMAGE_NAME:latest

echo "✅ 完成! 镜像已推送到:"
echo "   - $REPO_NAME/$IMAGE_NAME:$TAG"
echo "   - $REPO_NAME/$IMAGE_NAME:latest"