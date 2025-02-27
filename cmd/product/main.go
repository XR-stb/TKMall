package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TKMall/build/proto_gen/product"
	"TKMall/cmd/product/model"
	"TKMall/cmd/product/service"
	"TKMall/common/config"
	"TKMall/common/etcd"
	commonModel "TKMall/common/model"

	"TKMall/common/log"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	log.Init("config/log.yaml", "product")

	// 初始化配置
	config.InitConfig("product")

	// 连接数据库
	db, err := commonModel.InitGORM()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 注册到ETCD
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

	err = etcd.RegisterService(client, serviceName, fmt.Sprintf("localhost:%d", port), 10)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化Redis连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	// 启动gRPC服务
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	s := grpc.NewServer()
	product.RegisterProductCatalogServiceServer(s, &service.ProductCatalogServiceServer{
		DB:    db,
		Redis: redisClient,
	})

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		s.GracefulStop()
	}()

	log.Infof("商品服务启动成功，监听端口: %d", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
