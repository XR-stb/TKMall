# 获取当前sh文件的完整路径
SCRIPT_PATH=$(readlink -f "$0")
# 获取上级目录名
SHELL_DIR=$(dirname "$SCRIPT_PATH")
SCRIPTS_DIR=$(dirname "$SHELL_DIR")
# 获取到工作区路径
WORKSPACE_DIR=$(dirname "$SCRIPTS_DIR")

# protoc 路径
PROTC_BIN="${WORKSPACE_DIR}/tools/protobuf/bin/protoc"

eval $PROTC_BIN --version

# 所有proto文件的路径
PROTO_PATH="proto/"

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

# 生成pb.xxx
PROTOC_CMD="${PROTC_BIN} --proto_path=${PROTO_PATH} \
            --go_out=${GOLANG_OUT_PATH} \
            --go_opt=paths=source_relative \
            --go-grpc_out=${GOLANG_OUT_PATH} \
            --go-grpc_opt=paths=source_relative \
            ${PROTO_PATH}/*.proto"
eval $PROTOC_CMD
