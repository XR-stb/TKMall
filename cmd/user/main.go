package main

import (
	"TKMall/common/log"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	user "TKMall/build/proto_gen/user"
	"TKMall/common/config"
	"TKMall/common/etcd"

	"TKMall/cmd/user/model"
	service "TKMall/cmd/user/service"

	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	userEvents "TKMall/cmd/user/events"
	"TKMall/common/events"
)

func main() {
	log.Init("config/log.yaml", "user")
	config.InitConfig("user")

	serviceName := viper.GetString("server.name")
	port := viper.GetInt("server.port")
	etcdEndpoints := viper.GetStringSlice("etcd.endpoints")
	etcdDialTimeout := viper.GetDuration("etcd.dial_timeout") * time.Second

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 获取服务地址，优先使用环境变量中的POD_IP
	serviceHost := "localhost"
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		serviceHost = podIP
	}

	// 在Kubernetes环境中注册服务
	err = etcd.RegisterService(client, serviceName, fmt.Sprintf("%s:%d", serviceHost, port), 10)
	if err != nil {
		log.Fatal(err)
	}

	db, err := model.InitGORM()
	if err != nil {
		log.Fatalf("failed to initialize GORM: %v", err)
	}

	// 初始化雪花算法节点
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("failed to initialize snowflake node: %v", err)
	}

	// 初始化服务代理
	// 从环境变量获取auth服务地址，优先尝试多种可能的环境变量格式
	authServiceAddr := viper.GetString("auth_service.address")
	// 直接从环境变量获取
	if addr := os.Getenv("AUTH_SERVICE_ADDR"); addr != "" {
		log.Infof("使用环境变量地址 AUTH_SERVICE_ADDR: %s", addr)
		authServiceAddr = addr
	} else if host := os.Getenv("AUTH_SERVICE_SERVICE_HOST"); host != "" {
		// 从Kubernetes服务发现环境变量获取
		port := os.Getenv("AUTH_SERVICE_SERVICE_PORT")
		if port != "" {
			addr := fmt.Sprintf("%s:%s", host, port)
			log.Infof("使用Kubernetes服务发现环境变量地址: %s", addr)
			authServiceAddr = addr
		}
	} else {
		log.Infof("使用配置文件中的auth服务地址: %s", authServiceAddr)
	}

	log.Infof("最终使用的auth服务地址: %s", authServiceAddr)

	serviceEndpoints := map[string]string{
		"auth": authServiceAddr,
	}

	// 获取Redis地址，优先使用环境变量
	redisAddr := "localhost:6379"
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		redisAddr = addr
	}

	serviceProxy := proxy.NewGrpcProxy(serviceEndpoints, redisAddr)

	// 初始化事件总线
	kafkaBrokers := []string{"localhost:9092"} // 默认值

	// 从环境变量中读取Kafka地址
	if brokers := viper.GetStringSlice("kafka.brokers"); len(brokers) > 0 {
		kafkaBrokers = brokers
	}

	// 尝试从不同名称的环境变量中读取
	if os.Getenv("KAFKA_BROKERS") != "" {
		kafkaBrokers = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	} else if os.Getenv("KAFKA_ADDR") != "" {
		kafkaBrokers = strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	} else if os.Getenv("KAFKA_BOOTSTRAP_SERVERS") != "" {
		kafkaBrokers = strings.Split(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"), ",")
	} else if os.Getenv("KAFKA_BROKER_ADDRS") != "" {
		kafkaBrokers = strings.Split(os.Getenv("KAFKA_BROKER_ADDRS"), ",")
	}

	log.Infof("使用Kafka地址: %v", kafkaBrokers)
	eventBus, err := events.NewKafkaEventBus(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to initialize event bus: %v", err)
	}

	// 初始化事件处理器
	if err := userEvents.InitEventHandlers(eventBus); err != nil {
		log.Fatalf("Failed to initialize event handlers: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, &service.UserServiceServer{
		DB:       db,
		Node:     node,
		Proxy:    serviceProxy,
		EventBus: eventBus,
	})
	log.Infof("用户服务启动成功，监听端口: %d", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
