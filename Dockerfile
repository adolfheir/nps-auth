# 使用官方的Go镜像作为构建阶段的基础镜像
FROM golang:1.22-bullseye AS builder

# 设置工作目录
WORKDIR /app

# 将当前目录下的所有文件复制到容器中的工作目录
COPY . .

# 拉取依赖包
RUN go mod download

# 编译Go应用
RUN go build -o main . 

# 先用这个吧 后续弄交叉编译
FROM golang:1.22-bullseye  AS final

# # 将构建阶段的可执行文件复制到最终镜像中
COPY --from=builder /app/main /usr/local/bin/main

# 将工作目录设置为 /usr/local/bin
WORKDIR /usr/local/bin

# 设置时区
ENV TZ Asia/Shanghai

# 创建一个用于存放数据的目录
RUN mkdir -p /data

# 暴露应用运行的端口，例如30106
EXPOSE 30106

# 运行应用
CMD ["./main","server"]
