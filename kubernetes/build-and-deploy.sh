#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # 无颜色

# 检查是否安装了Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}请先安装Docker${NC}"
    exit 1
fi

# 检查是否安装了kubectl
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}请先安装kubectl${NC}"
    exit 1
fi

# 检查是否有Kubernetes集群
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}无法连接到Kubernetes集群${NC}"
    echo -e "${YELLOW}您在WSL中，可能需要安装并启动Minikube或Kind${NC}"
    echo -e "安装Minikube: https://minikube.sigs.k8s.io/docs/start/"
    echo -e "或使用Kind: https://kind.sigs.k8s.io/docs/user/quick-start/"
    exit 1
fi

# 设置镜像标签 (使用本地镜像，不需要推送到仓库)
TAG="latest"
# 使用本地镜像
USE_LOCAL_IMAGES=true
# 镜像仓库前缀
REGISTRY_PREFIX="tkmall"
# 所有服务列表
SERVICES="user product cart order payment checkout auth gateway"

# 显示用法信息
usage() {
    echo -e "${YELLOW}用法:${NC}"
    echo -e "  $0 [选项]"
    echo
    echo -e "${YELLOW}选项:${NC}"
    echo -e "  build             构建所有服务的Docker镜像"
    echo -e "  build-specific    构建指定服务的Docker镜像，例如：$0 build-specific gateway,user"
    echo -e "  deploy-infra      部署基础设施（MySQL, Redis, ETCD, Kafka）"
    echo -e "  deploy-services   部署所有微服务（会重新构建镜像）"
    echo -e "  deploy-services-only  部署所有微服务（不重新构建镜像）"
    echo -e "  deploy-specific   部署指定的微服务，例如：$0 deploy-specific gateway,user"
    echo -e "  deploy-all        部署所有组件"
    echo -e "  clean             清理所有部署的资源"
    echo -e "  help              显示此帮助信息"
}

# 修复Dockerfile中的Go版本问题
fix_dockerfile_go_version() {
    local service=$1
    
    # 获取正确的根目录路径
    local current_dir=$(pwd)
    local root_dir=""
    
    # 情况1：当前目录是TKMall
    if [[ "$(basename "$current_dir")" == "TKMall" ]]; then
        root_dir="$current_dir"
    # 情况2：当前目录是TKMall的子目录（如kubernetes）
    elif [[ "$(basename "$(dirname "$current_dir")")" == "TKMall" ]]; then
        root_dir="$(dirname "$current_dir")"
    # 情况3：当前目录包含TKMall子目录
    elif [ -d "$current_dir/TKMall" ]; then
        root_dir="$current_dir/TKMall"
    # 情况4：往上一级找TKMall
    elif [ -d "$(dirname "$current_dir")/TKMall" ]; then
        root_dir="$(dirname "$current_dir")/TKMall"
    else
        # 使用相对路径尝试推断
        root_dir=$(cd "$(dirname "$0")/.." 2>/dev/null && pwd)
    fi
    
    local dockerfile="${root_dir}/kubernetes/docker/Dockerfile.${service}"
    
    echo -e "${YELLOW}处理Dockerfile: $dockerfile${NC}"
    
    if [ ! -f "$dockerfile" ]; then
        echo -e "${RED}Dockerfile不存在: $dockerfile${NC}"
        return 1
    fi
    
    # 检查Dockerfile是否已经包含修复Go版本的命令
    if ! grep -q "RUN sed -i 's/go 1.23.4/go 1.23/g' go.mod" "$dockerfile"; then
        # 在"COPY go.mod go.sum ./"行后添加修复命令
        sed -i '/COPY go.mod go.sum .\//a # 修复go.mod中的版本格式问题\nRUN sed -i '\''s/go 1.23.4/go 1.23/g'\'' go.mod && cat go.mod' "$dockerfile"
        echo -e "${GREEN}已更新 $dockerfile 以修复Go版本格式问题${NC}"
    else
        echo -e "${YELLOW}$dockerfile 已包含修复Go版本格式的命令${NC}"
    fi
    
    # 检查是否已添加GOTOOLCHAIN环境变量
    if ! grep -q "ENV GOTOOLCHAIN=" "$dockerfile"; then
        # 在FROM行后添加GOTOOLCHAIN环境变量
        sed -i '/^FROM golang:/a ENV GOTOOLCHAIN=auto' "$dockerfile"
        echo -e "${GREEN}已添加GOTOOLCHAIN=auto环境变量到 $dockerfile ${NC}"
    else
        echo -e "${YELLOW}$dockerfile 已包含GOTOOLCHAIN环境变量${NC}"
    fi
    
    # 更新Go版本到1.22-alpine
    sed -i 's/FROM golang:1.20-alpine/FROM golang:1.22-alpine/g' "$dockerfile"
    sed -i 's/FROM golang:1.21-alpine/FROM golang:1.22-alpine/g' "$dockerfile"
    echo -e "${GREEN}已更新 $dockerfile 中的Go版本到1.22-alpine${NC}"

    # 设置时区和时间 - 完整版本
    if ! grep -q "RUN apk add --no-cache tzdata" "$dockerfile"; then
        sed -i '/^FROM golang:/a RUN apk add --no-cache tzdata\nENV TZ=Asia/Shanghai\nRUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone\nRUN date' "$dockerfile"
        echo -e "${GREEN}已添加完整的时区和时间设置到 $dockerfile ${NC}"
    else
        echo -e "${YELLOW}$dockerfile 已包含时区配置${NC}"
    fi
}

# 构建所有服务的Docker镜像
build_images() {
    echo -e "${GREEN}开始构建所有服务的Docker镜像...${NC}"
    
    # 确保根路径包含TKMall
    local current_dir=$(pwd)
    local root_dir=""
    
    # 情况1：当前目录是TKMall
    if [[ "$(basename "$current_dir")" == "TKMall" ]]; then
        root_dir="$current_dir"
    # 情况2：当前目录是TKMall的子目录（如kubernetes）
    elif [[ "$(basename "$(dirname "$current_dir")")" == "TKMall" ]]; then
        root_dir="$(dirname "$current_dir")"
    # 情况3：当前目录包含TKMall子目录
    elif [ -d "$current_dir/TKMall" ]; then
        root_dir="$current_dir/TKMall"
    # 情况4：往上一级找TKMall
    elif [ -d "$(dirname "$current_dir")/TKMall" ]; then
        root_dir="$(dirname "$current_dir")/TKMall"
    else
        echo -e "${RED}错误: 无法找到TKMall目录${NC}"
        echo -e "${YELLOW}当前目录: $current_dir${NC}"
        echo -e "${YELLOW}目录内容: $(ls -la)${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}推断出的根目录: $root_dir${NC}"
    cd "$root_dir"
    
    echo -e "${YELLOW}当前构建目录: $(pwd)${NC}"
    
    # 确认Docker文件目录存在
    if [ ! -d "$root_dir/kubernetes/docker" ]; then
        echo -e "${RED}错误: Docker文件目录不存在: $root_dir/kubernetes/docker${NC}"
        echo -e "${YELLOW}当前目录内容: $(ls -la)${NC}"
        echo -e "${YELLOW}kubernetes目录内容: $(ls -la kubernetes 2>/dev/null || echo '目录不存在')${NC}"
        exit 1
    fi
    
    # 构建各个服务的镜像
    for service in $SERVICES; do
        echo -e "${YELLOW}构建 $service 服务镜像...${NC}"
        
        # 确保Dockerfile已修复
        fix_dockerfile_go_version "$service"
        
        # 使用服务特定的Dockerfile
        local dockerfile="${root_dir}/kubernetes/docker/Dockerfile.${service}"
        if [ ! -f "$dockerfile" ]; then
            echo -e "${RED}Dockerfile不存在，跳过构建: $dockerfile${NC}"
            continue
        fi
        
        echo -e "${YELLOW}使用Dockerfile: $dockerfile${NC}"
        docker build -t ${REGISTRY_PREFIX}/${service}:${TAG} -f "$dockerfile" .
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}${service} 服务镜像构建成功${NC}"
        else
            echo -e "${RED}${service} 服务镜像构建失败${NC}"
            exit 1
        fi
    done
    
    echo -e "${GREEN}所有镜像构建完成${NC}"
}

# 构建指定服务的Docker镜像
build_specific_images() {
    local specific_services=$1
    echo -e "${GREEN}开始构建指定服务的Docker镜像: $specific_services ${NC}"
    
    # 确保根路径包含TKMall
    local current_dir=$(pwd)
    local root_dir=""
    
    # 情况1：当前目录是TKMall
    if [[ "$(basename "$current_dir")" == "TKMall" ]]; then
        root_dir="$current_dir"
    # 情况2：当前目录是TKMall的子目录（如kubernetes）
    elif [[ "$(basename "$(dirname "$current_dir")")" == "TKMall" ]]; then
        root_dir="$(dirname "$current_dir")"
    # 情况3：当前目录包含TKMall子目录
    elif [ -d "$current_dir/TKMall" ]; then
        root_dir="$current_dir/TKMall"
    # 情况4：往上一级找TKMall
    elif [ -d "$(dirname "$current_dir")/TKMall" ]; then
        root_dir="$(dirname "$current_dir")/TKMall"
    else
        # 使用相对路径尝试推断
        root_dir=$(cd "$(dirname "$0")/.." 2>/dev/null && pwd)
    fi
    
    echo -e "${YELLOW}推断出的根目录: $root_dir${NC}"
    cd "$root_dir"
    
    echo -e "${YELLOW}当前构建目录: $(pwd)${NC}"
    
    # 确认Docker文件目录存在
    if [ ! -d "$root_dir/kubernetes/docker" ]; then
        echo -e "${RED}错误: Docker文件目录不存在: $root_dir/kubernetes/docker${NC}"
        echo -e "${YELLOW}当前目录内容: $(ls -la)${NC}"
        echo -e "${YELLOW}kubernetes目录内容: $(ls -la kubernetes 2>/dev/null || echo '目录不存在')${NC}"
        exit 1
    fi
    
    # 将以逗号分隔的服务列表转换为数组
    IFS=',' read -r -a services_array <<< "$specific_services"
    
    # 构建各个指定服务的镜像
    for service in "${services_array[@]}"; do
        echo -e "${YELLOW}构建 $service 服务镜像...${NC}"
        
        # 确保Dockerfile已修复
        fix_dockerfile_go_version "$service"
        
        # 使用服务特定的Dockerfile
        local dockerfile="${root_dir}/kubernetes/docker/Dockerfile.${service}"
        if [ ! -f "$dockerfile" ]; then
            echo -e "${RED}Dockerfile不存在，跳过构建: $dockerfile${NC}"
            continue
        fi
        
        echo -e "${YELLOW}使用Dockerfile: $dockerfile${NC}"
        docker build -t ${REGISTRY_PREFIX}/${service}:${TAG} -f "$dockerfile" .
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}${service} 服务镜像构建成功${NC}"
        else
            echo -e "${RED}${service} 服务镜像构建失败${NC}"
            exit 1
        fi
    done
    
    echo -e "${GREEN}指定镜像构建完成${NC}"
}

# 部署基础设施
deploy_infra() {
    echo -e "${GREEN}开始部署基础设施...${NC}"
    
    # 使用绝对路径
    local script_dir=$(cd "$(dirname "$0")" && pwd)
    cd "$script_dir"
    
    echo -e "${YELLOW}当前部署目录: $(pwd)${NC}"
    
    # 检查manifests目录是否存在
    if [ ! -d "manifests" ]; then
        echo -e "${RED}错误: manifests目录不存在${NC}"
        echo -e "${YELLOW}当前目录: $(pwd)${NC}"
        echo -e "${YELLOW}目录内容: $(ls -la)${NC}"
        exit 1
    fi
    
    # 创建命名空间（如果不存在）
    kubectl create namespace tkmall 2>/dev/null || true
    
    # 部署MySQL
    echo -e "${YELLOW}部署MySQL...${NC}"
    local mysql_file="${script_dir}/manifests/mysql-deployment.yaml"
    echo -e "${YELLOW}使用部署文件: $mysql_file${NC}"
    kubectl apply -f "$mysql_file" -n tkmall
    
    # 部署Redis
    echo -e "${YELLOW}部署Redis...${NC}"
    local redis_file="${script_dir}/manifests/redis-deployment.yaml"
    echo -e "${YELLOW}使用部署文件: $redis_file${NC}"
    kubectl apply -f "$redis_file" -n tkmall
    
    # 部署ETCD
    echo -e "${YELLOW}部署ETCD...${NC}"
    local etcd_file="${script_dir}/manifests/etcd-deployment.yaml"
    echo -e "${YELLOW}使用部署文件: $etcd_file${NC}"
    kubectl apply -f "$etcd_file" -n tkmall
    
    # 部署Kafka
    echo -e "${YELLOW}部署Kafka...${NC}"
    local kafka_file="${script_dir}/manifests/kafka-deployment.yaml"
    echo -e "${YELLOW}使用部署文件: $kafka_file${NC}"
    kubectl apply -f "$kafka_file" -n tkmall
    
    echo -e "${GREEN}基础设施部署完成${NC}"
    echo -e "${YELLOW}等待基础设施启动...${NC}"
    
    # 等待基础设施部署就绪
    kubectl rollout status deployment/mysql -n tkmall
    kubectl rollout status deployment/redis -n tkmall
    kubectl rollout status deployment/etcd -n tkmall
    kubectl rollout status deployment/zookeeper -n tkmall
    kubectl rollout status deployment/kafka -n tkmall
    
    echo -e "${GREEN}基础设施已就绪${NC}"

    # 部署MySQL后等待它准备就绪，确保它完全启动
    echo -e "${YELLOW}等待MySQL完全就绪（额外等待30秒）...${NC}"
    kubectl rollout status deployment/mysql -n tkmall
    sleep 30
    
    # 检查MySQL Pod状态
    echo -e "${YELLOW}检查MySQL Pod状态:${NC}"
    kubectl get pods -n tkmall -l app=mysql
    
    # 检查MySQL服务状态
    echo -e "${YELLOW}检查MySQL服务状态:${NC}"
    kubectl get service mysql-service -n tkmall
    
    # 获取MySQL Pod名称
    local mysql_pod=$(kubectl get pods -n tkmall -l app=mysql -o jsonpath='{.items[0].metadata.name}')
    echo -e "${YELLOW}MySQL Pod名称: $mysql_pod${NC}"
    
    # 获取MySQL Root密码
    local mysql_root_pwd=$(kubectl get secret -n tkmall mysql-secret -o jsonpath="{.data.root-password}" 2>/dev/null | base64 --decode)
    if [ -z "$mysql_root_pwd" ]; then
        echo -e "${YELLOW}从密钥中未找到root密码，使用默认密码'root'${NC}"
        mysql_root_pwd="root" # 默认密码
    else
        echo -e "${YELLOW}从密钥中获取到root密码: ${mysql_root_pwd:0:1}*****${NC}"
    fi
    
    echo -e "${YELLOW}使用MySQL Root密码初始化数据库...${NC}"
    
    # 输出MySQL容器的环境变量，查看是否设置了密码
    echo -e "${YELLOW}MySQL容器环境变量:${NC}"
    kubectl exec -n tkmall $mysql_pod -- env | grep MYSQL

    # 先检查MySQL是否正常工作
    echo -e "${YELLOW}尝试连接MySQL使用密码: ${mysql_root_pwd:0:1}*****${NC}"
    if ! kubectl exec -i -n tkmall $mysql_pod -- mysql -u root -p${mysql_root_pwd} -e "SELECT 1;" &>/dev/null; then
        echo -e "${RED}错误: 无法连接到MySQL。检查密码是否正确。${NC}"
        
        # 尝试列出正在使用的身份验证插件
        echo -e "${YELLOW}尝试不带密码连接MySQL...${NC}"
        kubectl exec -i -n tkmall $mysql_pod -- mysql -u root --skip-password -e "SELECT 1;" || echo "无法无密码连接"
        
        echo -e "${YELLOW}尝试使用默认密码 'root'...${NC}"
        mysql_root_pwd="root"
        
        if ! kubectl exec -i -n tkmall $mysql_pod -- mysql -u root -p${mysql_root_pwd} -e "SELECT 1;" &>/dev/null; then
            echo -e "${RED}错误: 无法连接到MySQL，即使使用默认密码也不行。${NC}"
            
            echo -e "${YELLOW}查看MySQL日志以了解更多信息:${NC}"
            kubectl logs -n tkmall $mysql_pod
            
            echo -e "${RED}MySQL连接失败，退出部署。${NC}"
            exit 1
        fi
    fi

    # 查看当前的MySQL用户
    echo -e "${YELLOW}当前MySQL用户:${NC}"
    kubectl exec -i -n tkmall $mysql_pod -- mysql -u root -p${mysql_root_pwd} -e "SELECT User, Host FROM mysql.user;"

    # 初始化数据库和权限 - 修复用户创建和授权
    echo -e "${YELLOW}创建并授权数据库用户...${NC}"
    cat <<EOF | kubectl exec -i -n tkmall $mysql_pod -- mysql -u root -p${mysql_root_pwd}
DROP USER IF EXISTS 'tkmalluser'@'%';
CREATE USER 'tkmalluser'@'%' IDENTIFIED BY 'yourpassword';
CREATE DATABASE IF NOT EXISTS shop;
GRANT ALL PRIVILEGES ON shop.* TO 'tkmalluser'@'%';
FLUSH PRIVILEGES;
SELECT User, Host FROM mysql.user;
SHOW GRANTS FOR 'tkmalluser'@'%';
EOF

    # 创建或更新MySQL连接密钥
    echo -e "${YELLOW}创建MySQL连接密钥...${NC}"
    kubectl delete secret mysql-connect-secret -n tkmall 2>/dev/null || true
    kubectl create secret generic mysql-connect-secret \
        --from-literal=dsn="tkmalluser:yourpassword@tcp(mysql-service:3306)/shop?charset=utf8mb4&parseTime=True&loc=Local" \
        -n tkmall
    
    # 保留原来的MySQL root密码密钥
    kubectl delete secret mysql-secret -n tkmall 2>/dev/null || true
    kubectl create secret generic mysql-secret \
        --from-literal=root-password="${mysql_root_pwd}" \
        -n tkmall
    
    echo -e "${GREEN}数据库初始化完成${NC}"
    
    # 检查用户是否能正常连接
    echo -e "${YELLOW}测试tkmalluser用户连接...${NC}"
    if kubectl exec -i -n tkmall $mysql_pod -- mysql -u tkmalluser -pyourpassword -e "USE shop; SELECT 'Connection successful';" 2>/dev/null; then
        echo -e "${GREEN}数据库用户连接测试成功!${NC}"
    else
        echo -e "${RED}警告: 数据库用户连接测试失败，服务可能无法连接数据库${NC}"
        # 显示MySQL连接错误
        kubectl exec -i -n tkmall $mysql_pod -- mysql -u tkmalluser -pyourpassword -e "USE shop;"
    fi
}

# 部署微服务
deploy_services() {
    echo -e "${GREEN}开始部署微服务...${NC}"
    
    # 使用绝对路径
    local script_dir=$(cd "$(dirname "$0")" && pwd)
    cd "$script_dir"
    
    echo -e "${YELLOW}当前部署目录: $(pwd)${NC}"
    
    # 检查manifests目录是否存在
    if [ ! -d "manifests" ]; then
        echo -e "${RED}错误: manifests目录不存在${NC}"
        echo -e "${YELLOW}当前目录: $(pwd)${NC}"
        echo -e "${YELLOW}目录内容: $(ls -la)${NC}"
        exit 1
    fi
    
    # 检查部署文件是否存在
    missing_files=false
    for service in $SERVICES; do
        if [ ! -f "manifests/${service}-deployment.yaml" ]; then
            echo -e "${RED}错误: manifests/${service}-deployment.yaml 不存在${NC}"
            missing_files=true
        fi
    done
    
    if [ "$missing_files" = true ]; then
        echo -e "${RED}错误: 部分部署文件不存在，退出部署${NC}"
        exit 1
    fi
    
    # 创建命名空间（如果不存在）
    kubectl create namespace tkmall 2>/dev/null || true
    
    # 如果使用本地镜像，需要将本地镜像加载到Minikube/Kind中
    if [ "$USE_LOCAL_IMAGES" = true ]; then
        echo -e "${YELLOW}将本地镜像加载到Kubernetes集群中...${NC}"
        
        # 检测是否使用Minikube
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            echo -e "${YELLOW}检测到Minikube环境，使用minikube docker-env${NC}"
            eval $(minikube docker-env)
            
            # 获取正确的根目录路径（不改变当前目录）
            local current_dir=$(pwd)
            local root_dir=""
            
            # 情况1：当前目录是TKMall
            if [[ "$(basename "$current_dir")" == "TKMall" ]]; then
                root_dir="$current_dir"
            # 情况2：当前目录是TKMall的子目录（如kubernetes）
            elif [[ "$(basename "$(dirname "$current_dir")")" == "TKMall" ]]; then
                root_dir="$(dirname "$current_dir")"
            # 情况3：当前目录包含TKMall子目录
            elif [ -d "$current_dir/TKMall" ]; then
                root_dir="$current_dir/TKMall"
            # 情况4：往上一级找TKMall
            elif [ -d "$(dirname "$current_dir")/TKMall" ]; then
                root_dir="$(dirname "$current_dir")/TKMall"
            else
                # 使用script_dir的上级目录
                root_dir=$(dirname "$script_dir")
            fi
            
            echo -e "${YELLOW}使用根目录: $root_dir 进行构建${NC}"
            
            # 在当前环境中重新构建镜像
            local current_pwd=$(pwd)
            cd "$root_dir"
            build_images
            cd "$current_pwd"  # 返回原来的目录
        
        # 检测是否使用Kind
        elif command -v kind &> /dev/null && kind get clusters &> /dev/null; then
            echo -e "${YELLOW}检测到Kind环境，使用kind load${NC}"
            for service in $SERVICES; do
                kind load docker-image ${REGISTRY_PREFIX}/${service}:${TAG}
            done
        else
            echo -e "${YELLOW}使用本地Docker环境${NC}"
            # 修改部署文件以使用imagePullPolicy: Never
            for deployment in manifests/*-deployment.yaml; do
                sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Never/g' $deployment
            done
        fi
    fi
    
    # 部署各个微服务
    for service in $SERVICES; do
        echo -e "${YELLOW}部署 $service 服务...${NC}"
        local deploy_file="${script_dir}/manifests/${service}-deployment.yaml"
        echo -e "${YELLOW}使用部署文件: $deploy_file${NC}"
        kubectl apply -f "$deploy_file" -n tkmall
    done
    
    echo -e "${GREEN}所有微服务部署完成${NC}"
    echo -e "${YELLOW}等待微服务启动...${NC}"
    
    # 等待微服务部署就绪
    for service in $SERVICES; do
        kubectl rollout status deployment/${service}-service -n tkmall
    done
    
    echo -e "${GREEN}所有微服务已就绪${NC}"
    
    # 显示服务信息
    echo -e "${YELLOW}服务信息:${NC}"
    kubectl get services -n tkmall
    
    # 获取网关服务的外部IP (如果是LoadBalancer类型)
    if kubectl get service gateway-service -n tkmall | grep -q LoadBalancer; then
        echo -e "${YELLOW}获取网关服务的访问地址...${NC}"
        
        # 在Minikube中使用特殊命令获取URL
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            echo -e "${GREEN}网关访问地址: $(minikube service gateway-service -n tkmall --url)${NC}"
        else
            # 等待LoadBalancer获取外部IP
            EXTERNAL_IP=""
            while [ -z "$EXTERNAL_IP" ]; do
                EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
                if [ -z "$EXTERNAL_IP" ]; then
                    EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
                fi
                
                if [ -z "$EXTERNAL_IP" ]; then
                    echo -e "${YELLOW}等待网关服务获取外部IP...${NC}"
                    sleep 10
                fi
            done
            
            PORT=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.spec.ports[0].port}')
            echo -e "${GREEN}网关访问地址: http://$EXTERNAL_IP:$PORT${NC}"
        fi
    else
        echo -e "${YELLOW}网关服务不是LoadBalancer类型，需要使用port-forward访问:${NC}"
        echo -e "${GREEN}kubectl port-forward -n tkmall service/gateway-service 8080:8080${NC}"
    fi
}

# 部署微服务（不重新构建）
deploy_services_only() {
    echo -e "${GREEN}开始部署微服务（不重新构建）...${NC}"
    
    # 使用绝对路径
    local script_dir=$(cd "$(dirname "$0")" && pwd)
    cd "$script_dir"
    
    echo -e "${YELLOW}当前部署目录: $(pwd)${NC}"
    
    # 检查manifests目录是否存在
    if [ ! -d "manifests" ]; then
        echo -e "${RED}错误: manifests目录不存在${NC}"
        echo -e "${YELLOW}当前目录: $(pwd)${NC}"
        echo -e "${YELLOW}目录内容: $(ls -la)${NC}"
        exit 1
    fi
    
    # 检查部署文件是否存在
    missing_files=false
    for service in $SERVICES; do
        if [ ! -f "manifests/${service}-deployment.yaml" ]; then
            echo -e "${RED}错误: manifests/${service}-deployment.yaml 不存在${NC}"
            missing_files=true
        fi
    done
    
    if [ "$missing_files" = true ]; then
        echo -e "${RED}错误: 部分部署文件不存在，退出部署${NC}"
        exit 1
    fi
    
    # 创建命名空间（如果不存在）
    kubectl create namespace tkmall 2>/dev/null || true
    
    # 部署各个微服务
    for service in $SERVICES; do
        echo -e "${YELLOW}部署 $service 服务...${NC}"
        local deploy_file="${script_dir}/manifests/${service}-deployment.yaml"
        echo -e "${YELLOW}使用部署文件: $deploy_file${NC}"
        kubectl apply -f "$deploy_file" -n tkmall
    done
    
    echo -e "${GREEN}所有微服务部署完成${NC}"
    echo -e "${YELLOW}等待微服务启动...${NC}"
    
    # 等待微服务部署就绪
    for service in $SERVICES; do
        kubectl rollout status deployment/${service}-service -n tkmall
    done
    
    echo -e "${GREEN}所有微服务已就绪${NC}"
    
    # 显示服务信息
    echo -e "${YELLOW}服务信息:${NC}"
    kubectl get services -n tkmall
    
    # 获取网关服务的外部IP (如果是LoadBalancer类型)
    if kubectl get service gateway-service -n tkmall | grep -q LoadBalancer; then
        echo -e "${YELLOW}获取网关服务的访问地址...${NC}"
        
        # 在Minikube中使用特殊命令获取URL
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            echo -e "${GREEN}网关访问地址: $(minikube service gateway-service -n tkmall --url)${NC}"
        else
            # 等待LoadBalancer获取外部IP
            EXTERNAL_IP=""
            while [ -z "$EXTERNAL_IP" ]; do
                EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
                if [ -z "$EXTERNAL_IP" ]; then
                    EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
                fi
                
                if [ -z "$EXTERNAL_IP" ]; then
                    echo -e "${YELLOW}等待网关服务获取外部IP...${NC}"
                    sleep 10
                fi
            done
            
            PORT=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.spec.ports[0].port}')
            echo -e "${GREEN}网关访问地址: http://$EXTERNAL_IP:$PORT${NC}"
        fi
    else
        echo -e "${YELLOW}网关服务不是LoadBalancer类型，需要使用port-forward访问:${NC}"
        echo -e "${GREEN}kubectl port-forward -n tkmall service/gateway-service 8080:8080${NC}"
    fi
}

# 部署指定的微服务
deploy_specific_services() {
    local specific_services=$1
    echo -e "${GREEN}开始部署指定的微服务: $specific_services ${NC}"
    
    # 使用绝对路径
    local script_dir=$(cd "$(dirname "$0")" && pwd)
    cd "$script_dir"
    
    echo -e "${YELLOW}当前部署目录: $(pwd)${NC}"
    
    # 检查manifests目录是否存在
    if [ ! -d "manifests" ]; then
        echo -e "${RED}错误: manifests目录不存在${NC}"
        echo -e "${YELLOW}当前目录: $(pwd)${NC}"
        echo -e "${YELLOW}目录内容: $(ls -la)${NC}"
        exit 1
    fi
    
    # 将以逗号分隔的服务列表转换为数组
    IFS=',' read -r -a services_array <<< "$specific_services"
    
    # 检查部署文件是否存在
    missing_files=false
    for service in "${services_array[@]}"; do
        if [ ! -f "manifests/${service}-deployment.yaml" ]; then
            echo -e "${RED}错误: manifests/${service}-deployment.yaml 不存在${NC}"
            missing_files=true
        fi
    done
    
    if [ "$missing_files" = true ]; then
        echo -e "${RED}错误: 部分部署文件不存在，退出部署${NC}"
        exit 1
    fi
    
    # 创建命名空间（如果不存在）
    kubectl create namespace tkmall 2>/dev/null || true
    
    # 如果使用本地镜像，需要将本地镜像加载到Minikube/Kind中
    if [ "$USE_LOCAL_IMAGES" = true ]; then
        echo -e "${YELLOW}将指定的本地镜像加载到Kubernetes集群中...${NC}"
        
        # 检测是否使用Minikube
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            echo -e "${YELLOW}检测到Minikube环境，使用minikube docker-env${NC}"
            eval $(minikube docker-env)
            
            # 获取正确的根目录路径（不改变当前目录）
            local current_dir=$(pwd)
            local root_dir=""
            
            # 情况1：当前目录是TKMall
            if [[ "$(basename "$current_dir")" == "TKMall" ]]; then
                root_dir="$current_dir"
            # 情况2：当前目录是TKMall的子目录（如kubernetes）
            elif [[ "$(basename "$(dirname "$current_dir")")" == "TKMall" ]]; then
                root_dir="$(dirname "$current_dir")"
            # 情况3：当前目录包含TKMall子目录
            elif [ -d "$current_dir/TKMall" ]; then
                root_dir="$current_dir/TKMall"
            # 情况4：往上一级找TKMall
            elif [ -d "$(dirname "$current_dir")/TKMall" ]; then
                root_dir="$(dirname "$current_dir")/TKMall"
            else
                # 使用script_dir的上级目录
                root_dir=$(dirname "$script_dir")
            fi
            
            echo -e "${YELLOW}使用根目录: $root_dir 进行构建${NC}"
            
            # 在当前环境中重新构建指定镜像
            local current_pwd=$(pwd)
            cd "$root_dir"
            build_specific_images "$specific_services"
            cd "$current_pwd"  # 返回原来的目录
        
        # 检测是否使用Kind
        elif command -v kind &> /dev/null && kind get clusters &> /dev/null; then
            echo -e "${YELLOW}检测到Kind环境，使用kind load${NC}"
            for service in "${services_array[@]}"; do
                kind load docker-image ${REGISTRY_PREFIX}/${service}:${TAG}
            done
        else
            echo -e "${YELLOW}使用本地Docker环境${NC}"
            # 修改指定的部署文件以使用imagePullPolicy: Never
            for service in "${services_array[@]}"; do
                local deploy_file="manifests/${service}-deployment.yaml"
                if [ -f "$deploy_file" ]; then
                    sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Never/g' "$deploy_file"
                fi
            done
        fi
    fi
    
    # 部署各个指定的微服务
    for service in "${services_array[@]}"; do
        echo -e "${YELLOW}部署 $service 服务...${NC}"
        local deploy_file="${script_dir}/manifests/${service}-deployment.yaml"
        echo -e "${YELLOW}使用部署文件: $deploy_file${NC}"
        kubectl apply -f "$deploy_file" -n tkmall
    done
    
    echo -e "${GREEN}指定的微服务部署完成${NC}"
    echo -e "${YELLOW}等待微服务启动...${NC}"
    
    # 等待指定的微服务部署就绪
    for service in "${services_array[@]}"; do
        kubectl rollout status deployment/${service}-service -n tkmall
    done
    
    echo -e "${GREEN}所有指定的微服务已就绪${NC}"
    
    # 显示服务信息
    echo -e "${YELLOW}服务信息:${NC}"
    kubectl get services -n tkmall
    
    # 获取网关服务的外部IP (如果是LoadBalancer类型，且gateway在部署列表中)
    if [[ "$specific_services" == *"gateway"* ]] && kubectl get service gateway-service -n tkmall | grep -q LoadBalancer; then
        echo -e "${YELLOW}获取网关服务的访问地址...${NC}"
        
        # 在Minikube中使用特殊命令获取URL
        if command -v minikube &> /dev/null && minikube status &> /dev/null; then
            echo -e "${GREEN}网关访问地址: $(minikube service gateway-service -n tkmall --url)${NC}"
        else
            # 等待LoadBalancer获取外部IP
            EXTERNAL_IP=""
            while [ -z "$EXTERNAL_IP" ]; do
                EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
                if [ -z "$EXTERNAL_IP" ]; then
                    EXTERNAL_IP=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
                fi
                
                if [ -z "$EXTERNAL_IP" ]; then
                    echo -e "${YELLOW}等待网关服务获取外部IP...${NC}"
                    sleep 10
                fi
            done
            
            PORT=$(kubectl get service gateway-service -n tkmall -o jsonpath='{.spec.ports[0].port}')
            echo -e "${GREEN}网关访问地址: http://$EXTERNAL_IP:$PORT${NC}"
        fi
    elif [[ "$specific_services" == *"gateway"* ]]; then
        echo -e "${YELLOW}网关服务不是LoadBalancer类型，需要使用port-forward访问:${NC}"
        echo -e "${GREEN}kubectl port-forward -n tkmall service/gateway-service 8080:8080${NC}"
    fi
}

# 清理所有部署的资源
clean() {
    echo -e "${YELLOW}清理所有部署的资源...${NC}"
    
    # 删除命名空间（会删除所有相关资源）
    kubectl delete namespace tkmall
    
    echo -e "${GREEN}清理完成${NC}"
}

# 根据命令行参数执行相应的操作
case "$1" in
    build)
        build_images
        ;;
    build-specific)
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定要构建的服务，以逗号分隔${NC}"
            echo -e "${YELLOW}例如: $0 build-specific gateway,user${NC}"
            exit 1
        fi
        build_specific_images "$2"
        ;;
    deploy-infra)
        deploy_infra
        ;;
    deploy-services)
        deploy_services
        ;;
    deploy-services-only)
        deploy_services_only
        ;;
    deploy-specific)
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定要部署的服务，以逗号分隔${NC}"
            echo -e "${YELLOW}例如: $0 deploy-specific gateway,user${NC}"
            exit 1
        fi
        deploy_specific_services "$2"
        ;;
    deploy-all)
        deploy_infra
        deploy_services
        ;;
    clean)
        clean
        ;;
    help)
        usage
        ;;
    *)
        usage
        exit 1
        ;;
esac

exit 0 