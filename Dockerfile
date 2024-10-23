# 使用官方的Go镜像作为构建阶段的基础镜像
FROM golang:1.22.8 AS builder

# 设置工作目录
WORKDIR /app

# 将当前目录下的所有文件复制到容器中的工作目录
COPY . .

# 拉取依赖包
RUN go mod download

# 编译Go应用
RUN go build -o nps-auth .

# 使用一个更小的基础镜像，例如Alpine
FROM alpine:latest


# 将构建阶段的可执行文件复制到最终镜像中
COPY --from=builder /app/nps-auth /usr/local/bin/nps-auth

# 将工作目录设置为 /usr/local/bin
WORKDIR /usr/local/bin

# 暴露应用运行的端口，例如30106
EXPOSE 30106

# 运行应用
CMD ["nps-auth server"]
