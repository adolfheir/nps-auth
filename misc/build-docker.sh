

# 构建Go应用的Docker镜像
export IMAGE_NAME="ihouqi-docker.pkg.coding.net/fe/image/nps-auth:dev-1.0.0"
export CI_PIPELINE_ID="1"
export IMAGE_NAME="${IMAGE_NAME}-${CI_PIPELINE_ID}"

# 使用当前目录构建镜像
docker build   -t $IMAGE_NAME -f ./Dockerfile .
# docker push $IMAGE_NAME

echo "Docker镜像 $IMAGE_NAME 构建完成。"


