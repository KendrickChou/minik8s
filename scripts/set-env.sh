#!/bin/bash

if [ -z "$1" ]
  then
    echo "Need to input weave download address."
    exit 0
fi

sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list

apt-get update

curl -fsSL https://get.docker.com -o get-docker.sh

sh get-docker.sh

wget https://go.dev/dl/go1.18.2.linux-amd64.tar.gz

tar -C /usr/local -xzf go1.18.2.linux-amd64.tar.gz

echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

source ~/.bashrc

sudo curl -L $1 -o /usr/local/bin/weave

chmod a+x /usr/local/bin/weave

weave launch

go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct