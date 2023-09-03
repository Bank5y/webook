# 基础镜像
FROM ubuntu:20.04
# 将编译后的打包进来镜像，放到工作目录 /app
COPY webook /app/webook
WORKDIR /app

#CMD是执行命令
#最佳
ENTRYPOINT ["/app/webook"]
