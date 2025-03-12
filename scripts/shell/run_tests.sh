#!/bin/bash

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=============================${NC}"
echo -e "${YELLOW}开始运行TKMall单元测试套件${NC}"
echo -e "${YELLOW}=============================${NC}"

# 安装测试依赖
echo -e "${YELLOW}安装测试依赖...${NC}"
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/DATA-DOG/go-sqlmock

# 服务测试列表
services=(
    "cmd/auth/service"
    "cmd/user/service"
    "cmd/product/service"
    "cmd/cart/service"
    "cmd/payment/service"
    "cmd/gateway/middleware"
)

# 统计
total=0
passed=0
failed=0

run_test() {
    local package=$1
    echo -e "${YELLOW}运行测试: ${package}${NC}"
    
    # 增加详细输出的测试命令
    if go test -v ./${package} -count=1; then
        echo -e "${GREEN}✓ 测试通过: ${package}${NC}"
        ((passed++))
    else
        echo -e "${RED}✗ 测试失败: ${package}${NC}"
        ((failed++))
    fi
    ((total++))
    echo ""
}

# 运行各个服务的测试
for service in "${services[@]}"; do
    run_test "$service"
done

# 计算覆盖率
echo -e "${YELLOW}计算测试覆盖率...${NC}"
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# 打印测试摘要
echo -e "${YELLOW}=============================${NC}"
echo -e "${YELLOW}测试摘要${NC}"
echo -e "${YELLOW}=============================${NC}"
echo -e "总计: ${total} 服务"
echo -e "${GREEN}通过: ${passed}${NC}"
echo -e "${RED}失败: ${failed}${NC}"

# 设置退出状态
if [ $failed -gt 0 ]; then
    exit 1
else
    echo -e "${GREEN}所有测试通过!${NC}"
    exit 0
fi 