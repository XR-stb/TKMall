#!/bin/bash

# 获取当前sh文件的完整路径
SCRIPT_PATH=$(readlink -f "$0")
# 获取上级目录名
SHELL_DIR=$(dirname "$SCRIPT_PATH")
SCRIPTS_DIR=$(dirname "$SHELL_DIR")
# 获取到工作区路径
WORKSPACE_DIR=$(dirname "$SCRIPTS_DIR")

# 要关闭的端口列表
PORTS=(8080 50051 50052)

for PORT in "${PORTS[@]}"; do
    PID=$(lsof -ti:$PORT)
    if [ ! -z "$PID" ]; then
        PROCESS_NAME=$(ps -p $PID -o comm=)
        echo "Killing process on port $PORT (PID: $PID | NAME: $PROCESS_NAME)"
        kill -9 $PID
    else
        echo "No process found on port $PORT"
    fi
done 