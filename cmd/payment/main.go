package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TKMall/build/proto_gen/payment"
	"TKMall/cmd/payment/model"
	"TKMall/cmd/payment/service"
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
	log.Init("config/log.yaml", "payment")

	// 初始化配置
	config.InitConfig("payment")

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
	node, err := snowflake.NewNode(5) // 为支付服务使用节点ID 5
	if err != nil {
		log.Fatalf("初始化雪花ID节点失败: %v", err)
	}

	// 初始化服务代理
	serviceEndpoints := map[string]string{
		"order": viper.GetString("order_service.address"),
	}
	serviceProxy := proxy.NewGrpcProxy(serviceEndpoints, viper.GetString("redis.addr"))

	// 创建gRPC服务器
	server := grpc.NewServer()

	// 初始化支付服务
	paymentService := &service.PaymentServiceServer{
		DB:    db,
		Redis: rdb,
		Node:  node,
		Proxy: serviceProxy,
	}

	// 注册支付服务
	payment.RegisterPaymentServiceServer(server, paymentService)

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
	err = etcd.RegisterService(cli, serviceName, fmt.Sprintf("localhost:%d", port), 10)
	if err != nil {
		log.Fatalf("服务注册到etcd失败: %v", err)
	}

	// 启动gRPC服务
	go func() {
		log.Infof("支付服务启动成功，监听端口: %d", port)
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
