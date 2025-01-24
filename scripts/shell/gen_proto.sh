#!/bin/bash

# 获取当前sh文件的完整路径
SCRIPT_PATH=$(readlink -f "$0")
# 获取上级目录名
SHELL_DIR=$(dirname "$SCRIPT_PATH")
SCRIPTS_DIR=$(dirname "$SHELL_DIR")
# 获取到工作区路径
WORKSPACE_DIR=$(dirname "$SCRIPTS_DIR")

# protoc 路径
PROTOC_BIN="${WORKSPACE_DIR}/tools/protobuf/bin/protoc"

eval $PROTOC_BIN --version

# 所有proto文件的路径
PROTO_PATH="${WORKSPACE_DIR}/proto"

# go 文件pb.go输出 路径
GOLANG_OUT_PATH="${WORKSPACE_DIR}/build/proto_gen"

# 清除 Go 输出目录下的文件
if [ -d "$GOLANG_OUT_PATH" ]; then
    echo "Clearing files in $GOLANG_OUT_PATH"
    rm -rf "${GOLANG_OUT_PATH:?}"/*
fi

# 检查并创建 Golang 输出路径
if [ ! -d "$GOLANG_OUT_PATH" ]; then
    echo "Golang Output path does not exist. Creating..."
    mkdir -p "$GOLANG_OUT_PATH"
fi

# 递归处理 proto 文件
find "$PROTO_PATH" -name "*.proto" | while read -r proto_file; do
    # 获取 proto 文件所在目录相对于 proto 根目录的相对路径
    relative_dir=$(dirname "${proto_file#$PROTO_PATH/}")
    # 创建相应的输出目录
    mkdir -p "${GOLANG_OUT_PATH}/${relative_dir}"
    # 生成 pb.go 文件
    ${PROTOC_BIN} --proto_path="${PROTO_PATH}" \
                  --go_out="${GOLANG_OUT_PATH}/" \
                  --go_opt=paths=source_relative \
                  --go-grpc_out="${GOLANG_OUT_PATH}/" \
                  --go-grpc_opt=paths=source_relative \
                  "$proto_file"
done
