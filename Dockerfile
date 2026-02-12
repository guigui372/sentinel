# --- 第一阶段：构建 (Builder) ---
FROM golang:alpine AS builder

WORKDIR /app

# 1. 复制依赖描述文件
COPY go.mod go.sum ./
RUN go mod download

# 2. 复制所有代码
COPY . .

# 3. 编译！
RUN CGO_ENABLED=0 GOOS=linux go build -o sentinel ./cmd/sentinel


# --- 第二阶段：运行 (Runner) ---
FROM alpine:latest
WORKDIR /root/

# 从第一阶段把编译好的二进制文件拿过来
COPY --from=builder /app/sentinel .

# 启动
CMD ["./sentinel"]
