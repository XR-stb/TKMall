package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/checkout"
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	"TKMall/cmd/checkout/service"
	"TKMall/common/config"
	"TKMall/common/etcd"
	"TKMall/common/log"
	commonModel "TKMall/common/model"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	log.Init("config/log.yaml", "checkout")

	// 初始化配置
	config.InitConfig("checkout")

	// 连接数据库
	db, err := commonModel.InitGORM()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	// 创建雪花ID节点
	node, err := snowflake.NewNode(6) // 为结账服务使用节点ID 6
	if err != nil {
		log.Fatalf("初始化雪花ID节点失败: %v", err)
	}

	// 创建gRPC客户端连接到订单服务
	orderConn, err := grpc.Dial(
		viper.GetString("order_service.address"),
		grpc.WithInsecure(), // 生产环境应该使用TLS
	)
	if err != nil {
		log.Fatalf("连接订单服务失败: %v", err)
	}
	defer orderConn.Close()
	orderClient := order.NewOrderServiceClient(orderConn)

	// 创建gRPC客户端连接到支付服务
	paymentConn, err := grpc.Dial(
		viper.GetString("payment_service.address"),
		grpc.WithInsecure(), // 生产环境应该使用TLS
	)
	if err != nil {
		log.Fatalf("连接支付服务失败: %v", err)
	}
	defer paymentConn.Close()
	paymentClient := payment.NewPaymentServiceClient(paymentConn)

	// 创建gRPC客户端连接到购物车服务
	cartConn, err := grpc.Dial(
		viper.GetString("cart_service.address"),
		grpc.WithInsecure(), // 生产环境应该使用TLS
	)
	if err != nil {
		log.Fatalf("连接购物车服务失败: %v", err)
	}
	defer cartConn.Close()
	cartClient := cart.NewCartServiceClient(cartConn)

	// 创建gRPC服务器
	server := grpc.NewServer()

	// 初始化结账服务
	checkoutService := &service.CheckoutServiceServer{
		DB:             db,
		Redis:          rdb,
		Node:           node,
		OrderService:   orderClient,
		PaymentService: paymentClient,
		CartService:    cartClient,
	}

	// 注册结账服务
	checkout.RegisterCheckoutServiceServer(server, checkoutService)

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
		log.Infof("结账服务启动在端口 %d", port)
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
