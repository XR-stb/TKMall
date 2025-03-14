package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TKMall/build/proto_gen/cart"
	"TKMall/cmd/cart/model"
	"TKMall/cmd/cart/service"
	"TKMall/common/config"
	"TKMall/common/etcd"
	"TKMall/common/log"
	commonModel "TKMall/common/model"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	log.Init("config/log.yaml", "cart")

	// 初始化配置
	config.InitConfig("cart")

	// 连接数据库
	db, err := commonModel.InitGORM()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	// 创建雪花ID节点
	node, err := snowflake.NewNode(3) // 为购物车服务使用节点ID 3
	if err != nil {
		log.Fatalf("初始化雪花ID节点失败: %v", err)
	}

	// 初始化服务代理
	// 从环境变量获取product服务地址，如果不存在则使用配置文件
	productServiceAddr := viper.GetString("product_service.address")
	if addr := os.Getenv("PRODUCT_SERVICE_ADDR"); addr != "" {
		log.Infof("使用环境变量地址 PRODUCT_SERVICE_ADDR: %s", addr)
		productServiceAddr = addr
	} else {
		log.Infof("使用配置文件中的product服务地址: %s", productServiceAddr)
	}

	serviceEndpoints := map[string]string{
		// 当购物车服务需要调用其他服务时，在这里添加
		"product": productServiceAddr,
	}

	// 获取Redis地址，优先使用环境变量
	redisAddr := viper.GetString("redis.addr")
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		log.Infof("使用环境变量地址 REDIS_ADDR: %s", addr)
		redisAddr = addr
	}

	serviceProxy := proxy.NewGrpcProxy(serviceEndpoints, redisAddr)

	// 创建gRPC服务器
	server := grpc.NewServer()

	// 初始化购物车服务
	cartService := &service.CartServiceServer{
		DB:    db,
		Redis: rdb,
		Node:  node,
		Proxy: serviceProxy,
	}

	// 注册购物车服务
	cart.RegisterCartServiceServer(server, cartService)

	// 获取服务配置
	port := viper.GetInt("server.port")
	serviceName := viper.GetString("server.name")

	// 启动gRPC服务
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("无法监听端口: %v", err)
	}

	// 服务注册到etcd
	etcdEndpoints := viper.GetStringSlice("etcd.endpoints")
	etcdDialTimeout := viper.GetDuration("etcd.dial_timeout") * time.Second

	// 创建etcd客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		log.Fatalf("连接etcd失败: %v", err)
	}
	defer cli.Close()

	// 注册服务
	serviceKey := fmt.Sprintf("/services/%s", serviceName)

	// 获取服务地址，优先使用环境变量中的POD_IP
	serviceHost := "localhost"
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		serviceHost = podIP
		log.Infof("使用POD_IP作为服务地址: %s", serviceHost)
	} else {
		log.Infof("使用localhost作为服务地址")
	}

	serviceValue := fmt.Sprintf("%s:%d", serviceHost, port)

	// 注册服务到etcd，设置TTL为10秒
	err = etcd.RegisterService(cli, serviceKey, serviceValue, 10)
	if err != nil {
		log.Fatalf("服务注册到etcd失败: %v", err)
	}

	// 启动gRPC服务
	go func() {
		log.Infof("购物车服务启动成功，监听端口: %d", port)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务...")

	// 优雅地关闭服务
	server.GracefulStop()
	log.Info("服务已关闭")
}
