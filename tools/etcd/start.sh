#!/bin/bash

# Define the installation directory
INSTALL_DIR=$(dirname "$0")/bin

# 检查 etcd 是否已安装
if [ ! -f "${INSTALL_DIR}/etcd" ]; then
    echo "etcd 未安装，正在进行安装..."
    # 调用安装脚本
    $(dirname "$0")/install.sh
else
    # 检查 etcd 是否已经运行
    if ! pgrep -x "etcd" > /dev/null; then
        echo "正在启动 etcd..."
        # Start etcd
        ${INSTALL_DIR}/etcd --data-dir tools/etcd/bin
    else
        echo "etcd 已经在运行中"
    fi
fi