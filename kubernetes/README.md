# TKMall 本地Kubernetes部署指南

本指南将帮助您在WSL (Windows Subsystem for Linux) 环境中搭建本地Kubernetes集群并部署TKMall微服务系统。

## 目录结构

```
kubernetes/
├── docker/             # 所有服务的Dockerfile
├── manifests/          # Kubernetes部署清单文件
└── build-and-deploy.sh # 部署脚本
```

## 服务架构

TKMall微服务系统包含以下服务：

- **gateway**: API网关服务，统一对外提供RESTful API
- **user**: 用户服务，处理用户注册、登录等功能
- **auth**: 认证服务，处理身份验证和授权
- **product**: 商品服务，管理商品信息
- **cart**: 购物车服务，管理用户购物车
- **order**: 订单服务，处理订单相关逻辑
- **payment**: 支付服务，处理支付相关功能
- **checkout**: 结账服务，处理下单流程

以及以下基础设施：

- **MySQL**: 数据持久化存储
- **Redis**: 缓存和会话存储
- **ETCD**: 服务发现和配置管理
- **Kafka**: 消息队列，用于事件驱动通信

## 前提条件

- WSL2已安装并正常运行
- Docker Desktop已安装并启用WSL集成
- kubectl命令行工具已安装

## 选择本地Kubernetes环境

在WSL中，您有以下几种方式来设置本地Kubernetes环境：

### 1. 使用Docker Desktop自带的Kubernetes

这是最简单的方法，如果您已经安装了Docker Desktop：

1. 打开Docker Desktop设置
2. 切换到"Kubernetes"选项卡
3. 勾选"Enable Kubernetes"并点击"Apply & Restart"
4. 等待Kubernetes启动完成

验证：
```bash
kubectl cluster-info
```

### 2. 使用Minikube

如果您想要更接近真实集群的体验：

1. 安装Minikube：
```bash
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

2. 启动Minikube：
```bash
minikube start --driver=docker
```

验证：
```bash
minikube status
kubectl cluster-info
```

### 3. 使用Kind (Kubernetes IN Docker)

Kind是一个更轻量级的选择：

1. 安装Kind：
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.17.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

2. 创建集群：
```bash
kind create cluster --name tkmall
```

验证：
```bash
kind get clusters
kubectl cluster-info
```

## 部署TKMall

确保您的Kubernetes环境已经就绪后，可以使用我们提供的脚本部署TKMall：

### 构建镜像

```bash
cd kubernetes
./build-and-deploy.sh build
```

### 部署基础设施

```bash
./build-and-deploy.sh deploy-infra
```

### 部署微服务

```bash
./build-and-deploy.sh deploy-services
```

### 不重新构建直接部署微服务（快速部署）

如果您已经构建了镜像，不需要重新构建，可以使用以下命令快速部署：

```bash
./build-and-deploy.sh deploy-services-only
```

### 或者一次性部署所有组件

```bash
./build-and-deploy.sh deploy-all
```

## 访问服务

部署完成后，您可以使用以下命令查看所有服务：

```bash
kubectl get pods -n tkmall
kubectl get services -n tkmall
```

### 访问API网关

Gateway服务是对外的统一接入点。如果您使用的是支持LoadBalancer类型的环境（如云平台或Docker Desktop Kubernetes），部署脚本会输出访问地址。

如果您使用的是本地环境，可以使用端口转发访问Gateway服务：

```bash
kubectl port-forward -n tkmall service/gateway-service 8080:8080
```

然后通过 http://localhost:8080 访问API网关。

#### 从Windows主机访问WSL2中的服务

如果您在WSL2中运行Kubernetes，并希望从Windows主机（如通过浏览器或API测试工具）访问这些服务，需要注意以下几点：

1. 默认的端口转发只绑定到127.0.0.1，这意味着只能在WSL2内部访问
2. 要从Windows主机访问，您需要让端口转发绑定到所有网络接口，并使用WSL2的IP地址

**步骤：**

1. 获取您的WSL2实例IP地址：
   ```bash
   ip addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'
   ```

2. 使用以下命令进行端口转发，注意增加`--address 0.0.0.0`参数：
   ```bash
   kubectl port-forward --address 0.0.0.0 -n tkmall service/gateway-service 8080:8080
   ```

3. 现在您可以在Windows主机上使用WSL2的IP地址访问服务：
   ```
   http://<WSL2-IP-地址>:8080
   ```

4. 例如，如果您的WSL2 IP是172.19.62.71，则可以访问：
   ```
   http://172.19.62.71:8080
   ```

**注意：** WSL2的IP地址可能会在每次重启后更改。如果连接突然失败，请重新检查IP地址。

### 访问其他服务

如果您需要直接访问某个后端服务，也可以使用端口转发：

```bash
# 示例：转发user服务端口
kubectl port-forward -n tkmall service/user-service 50051:50051
```

## 配置设计说明

TKMall系统设计了一种灵活的配置机制，适应不同的开发和部署环境：

### 本地开发环境与Kubernetes环境配置区别

1. **配置文件默认值**：
   - 所有服务的配置文件（如`config.yaml`）中默认使用`localhost`作为服务地址
   - 这确保了在本地开发时，各服务可以直接通过localhost互相访问

2. **Kubernetes环境变量覆盖**：
   - 在Kubernetes部署中，通过环境变量覆盖默认配置
   - 每个服务的部署清单（`*-deployment.yaml`）中定义了适合集群内通信的环境变量

3. **配置优先级**：
   - 环境变量具有最高优先级，会覆盖配置文件中的设置
   - 如果环境变量未设置，则使用配置文件中的默认值

### 本地开发与测试

在本地开发时，您可以：
- 直接使用配置文件中的默认localhost地址运行服务
- 或手动设置环境变量以测试不同配置

### 修改默认配置

如需修改默认配置：
- 本地开发：直接编辑各服务的配置文件
- Kubernetes部署：编辑对应服务的`manifests/*-deployment.yaml`文件中的环境变量

## 服务发现与依赖管理的最佳实践

当前的微服务架构使用了配置文件和环境变量来管理服务间依赖关系，虽然这种方式简单直观，但在服务数量增长时会变得难以维护。以下是几种更好的微服务依赖管理方式：

### 1. 使用服务发现机制

我们已经在使用ETCD进行服务注册，可以更进一步完善服务发现机制：

- **完全依赖服务发现**：服务启动时自动向服务注册中心(ETCD)注册，其他服务通过查询注册中心获取依赖服务地址，无需硬编码或环境变量
- **实现方式**：
  ```go
  // 示例：通过服务发现获取服务地址
  func getServiceAddress(serviceName string) (string, error) {
      resp, err := etcdClient.Get(context.Background(), fmt.Sprintf("/services/%s", serviceName))
      if err != nil {
          return "", err
      }
      if len(resp.Kvs) == 0 {
          return "", fmt.Errorf("service not found")
      }
      return string(resp.Kvs[0].Value), nil
  }
  ```

### 2. 使用集中式配置管理

- **配置服务**：创建一个专门的配置服务，所有服务从这里获取配置，包括依赖服务地址
- **实时更新**：配置服务可以支持实时更新和配置热加载，避免重启服务
- **实现方式**：可以使用Spring Cloud Config、Apollo等开源配置管理系统

### 3. 使用服务网格(Service Mesh)

对于更复杂的系统，可以考虑引入服务网格技术：

- **Istio/Linkerd**：提供统一的服务发现、流量管理和安全策略
- **边车代理(Sidecar)**：每个服务容器旁运行一个代理容器，自动处理服务发现、负载均衡等
- **好处**：服务代码不需要关心依赖服务的具体地址，由服务网格自动处理

### 4. 创建通用的服务连接库

- **统一服务连接逻辑**：创建一个通用库，封装服务连接、服务发现、配置获取等逻辑
- **示例实现**：
  ```go
  package serviceconnector

  // ServiceConnector 负责处理服务连接和发现
  type ServiceConnector struct {
      etcdClient *clientv3.Client
      // 其他字段...
  }

  // GetServiceClient 返回服务客户端
  func (c *ServiceConnector) GetServiceClient(serviceName string) (interface{}, error) {
      // 1. 从服务发现中心获取地址
      // 2. 建立连接
      // 3. 返回对应的客户端
      // ...
  }
  ```

### 5. 使用Kubernetes DNS服务发现

利用Kubernetes提供的DNS服务发现机制：

- **固定域名格式**：`<service-name>.<namespace>.svc.cluster.local`
- **简化配置**：只需要知道服务名称，不需要具体IP地址
- **健康检查**：Kubernetes自动处理服务实例的健康检查和负载均衡

### 推荐方案

对于TKMall项目，建议先实施以下改进：

1. **增强ETCD服务发现**：完善现有的ETCD服务注册和发现机制
2. **创建统一的服务连接库**：封装服务连接逻辑，统一处理环境变量、配置文件和服务发现
3. **利用Kubernetes DNS**：在Kubernetes环境中直接使用服务名称进行连接

随着项目规模扩大，再考虑引入服务网格或更高级的配置管理系统。

## 清理资源

当您不再需要运行TKMall时，可以使用以下命令清理所有资源：

```bash
./build-and-deploy.sh clean
```

如果您使用的是Minikube或Kind，也可以直接删除整个集群：

```bash
# Minikube
minikube delete

# Kind
kind delete cluster --name tkmall
```

## 故障排除

### 镜像无法拉取

如果您在部署时遇到镜像拉取问题，这可能是因为我们使用了本地构建的镜像。请确保您已经运行了`build`命令，并且Kubernetes集群可以访问这些镜像。

对于不同的环境，解决方案不同：
- Docker Desktop Kubernetes：通常可以直接使用本地Docker镜像
- Minikube：需要使用`eval $(minikube docker-env)`将Docker环境指向Minikube
- Kind：需要使用`kind load docker-image`命令将镜像加载到Kind集群中

### Pod启动失败

如果Pod无法正常启动，请检查以下内容：

```bash
# 查看Pod状态
kubectl get pods -n tkmall

# 查看Pod详情
kubectl describe pod <pod-name> -n tkmall

# 查看Pod日志
kubectl logs <pod-name> -n tkmall
```

### 资源不足

在WSL环境中，默认的资源限制可能不足以运行完整的TKMall系统。您可能需要调整WSL的资源配置：

1. 在Windows中创建或编辑`.wslconfig`文件（位于`%UserProfile%\.wslconfig`）
2. 添加或修改以下配置：
```
[wsl2]
memory=8GB
processors=4
```
3. 在PowerShell中重启WSL：`wsl --shutdown` 