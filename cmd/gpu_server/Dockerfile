FROM ubuntu:latest
ENV VERSION 1.0
# 在容器根目录 创建一个 apps 目录
WORKDIR /apps
# 拷贝当前目录下 go_docker_demo1 可以执行文件
COPY ./main /apps/main
# 设置编码
ENV LANG C.UTF-8

CMD ["/bin/bash"]