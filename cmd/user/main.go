package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"TKMall/build/proto_gen/auth"
	user "TKMall/build/proto_gen/user"
	"TKMall/common/config"

	"TKMall/cmd/user/model"
	service "TKMall/cmd/user/service"

	"github.com/bwmarrin/snowflake"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	config.InitConfig("user")

	port := viper.GetInt("server.port")
	etcdEndpoints := viper.GetStringSlice("etcd.endpoints")
	etcdDialTimeout := viper.GetDuration("etcd.dial_timeout") * time.Second
	authServiceAddress := viper.GetString("auth_service.address")

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 连接认证中心
	conn, err := grpc.Dial(authServiceAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer conn.Close()

	authClient := auth.NewAuthServiceClient(conn)

	err = registerService(client, "user-service", fmt.Sprintf("localhost:%d", port), 10)
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, &service.UserServiceServer{
		Users:      make(map[string]*service.User),
		AuthClient: authClient,
		DB:         db,
		Node:       node,
	})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
