# 多阶段构建 Dockerfile for LLM Member Management System

# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata sqlite

# 设置时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' > /etc/timezone

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制静态文件和配置文件
COPY --from=builder /app/webroot ./webroot
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/.env.example ./.env.example

# 创建存储目录
RUN mkdir -p ./storage

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 启动应用
CMD ["./main"]