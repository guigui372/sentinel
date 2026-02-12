name: Build and Push Docker Image

on:
  push:
    branches: [ "main" ]
  workflow_dispatch:

jobs:
  # --- 任务 1: 构建并推送到 Docker Hub (CI) ---
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and Push
        run: |
          # 注意：这里假设你的 Dockerfile 在根目录
          docker build -t ${{ secrets.DOCKER_USERNAME }}/sentinel-monitor:latest .
          docker push ${{ secrets.DOCKER_USERNAME }}/sentinel-monitor:latest

  # --- 任务 2: 部署到服务器 (CD) ---
  deploy:
    runs-on: ubuntu-latest
    needs: build-and-push  # 关键！一定要等构建成功了再部署
    steps:
      - name: Deploy to Server
        uses: appleboy/ssh-action@master # 使用第三方插件连接 SSH
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_KEY }}
          # 下面是登录服务器后要执行的命令脚本
          script: |
            # 1. 拉取最新镜像
            docker pull ${{ secrets.DOCKER_USERNAME }}/sentinel-monitor:latest
            
            # 2. 停止并删除旧容器 (如果存在)
            # "|| true" 的意思是：如果容器不存在，不要报错，继续往下执行
            docker stop sentinel-app || true
            docker rm sentinel-app || true
            
            # 3. 启动新容器
            # -d: 后台运行
            # --name: 给容器起个固定的名字，方便下次停止
            # -p 8080:8080: 映射端口 (根据你的程序端口修改)
            docker run -d \
              --name sentinel-app \
              --restart always \
              -p 8080:8080 \
              ${{ secrets.DOCKER_USERNAME }}/sentinel-monitor:latest
