# =================== 构建阶段 ===================
FROM golang:1.25.3-alpine AS builder

WORKDIR /src

# 安装 Node.js、pnpm 和 CGO 依赖（gcc + musl-dev）
RUN apk add --no-cache nodejs npm bash gcc musl-dev
RUN npm install -g pnpm

# 设置时区（可选）
ENV TZ=Asia/Shanghai

# 复制 Go 模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制全部源码
COPY . .

# 构建前端
RUN cd web && pnpm install --frozen-lockfile && pnpm build --mode production

# 编译 Go 二进制
# 注意：这里必须 CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -ldflags="-w -s" -o ech0 ./main.go

# =================== 最终镜像 ===================
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（便于 HTTPS 请求）
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建数据和备份目录
RUN mkdir -p /app/data /app/backup

# 从 builder 复制二进制
COPY --from=builder /src/ech0 /app/ech0

# 确保可执行
RUN chmod +x /app/ech0

EXPOSE 6277 6278

ENTRYPOINT ["/app/ech0"]
CMD ["serve"]