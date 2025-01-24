#!/bin/bash

# 检查系统类型
if [ -f /etc/redhat-release ]; then
    OS="centos"
elif [ -f /etc/lsb-release ]; then
    OS="ubuntu"
else
    echo "Unsupported OS"
    exit 1
fi

# 检查并安装 wget
if ! command -v wget &> /dev/null; then
    if [ "$OS" == "centos" ]; then
        sudo yum install -y wget
    elif [ "$OS" == "ubuntu" ]; then
        sudo apt-get update
        sudo apt-get install -y wget
    fi
fi

# 检查并安装 unzip
if ! command -v unzip &> /dev/null; then
    if [ "$OS" == "centos" ]; then
        sudo yum install -y unzip
    elif [ "$OS" == "ubuntu" ]; then
        sudo apt-get update
        sudo apt-get install -y unzip
    fi
fi

# 下载并解压 protoc
PROTOC_VERSION="29.3"
PROTOC_ZIP="protoc-${PROTOC_VERSION}-linux-x86_64.zip"
PROTOC_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"

wget $PROTOC_URL
unzip $PROTOC_ZIP -d /data/workspace/TKMall/tools/protobuf

# 添加 protoc 到 PATH
PROTOC_PATH="/data/workspace/TKMall/tools/protobuf/bin"
if ! grep -q "$PROTOC_PATH" <<< "$PATH"; then
    echo "export PATH=\$PATH:$PROTOC_PATH" >> ~/.bashrc
    source ~/.bashrc
fi

# 验证 protoc 是否可用
protoc --version

# 清理下载的 zip 文件
rm $PROTOC_ZIP

echo "protoc 安装和配置环境变量完成"
