package main

import (
	"TKMall/common/log"
	"fmt"
	"net"
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

	err = etcd.RegisterService(client, serviceName, fmt.Sprintf("localhost:%d", port), 10)
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
	serviceEndpoints := map[string]string{
		"auth": viper.GetString("auth_service.address"),
	}
	serviceProxy := proxy.NewGrpcProxy(serviceEndpoints)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, &service.UserServiceServer{
		DB:    db,
		Node:  node,
		Proxy: serviceProxy,
	})
	log.Infof("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
