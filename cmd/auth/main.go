package main

import (
	"TKMall/common/log"
	"fmt"
	"net"
	"os"
	"time"

	"TKMall/build/proto_gen/auth"
	"TKMall/cmd/auth/service"
	"TKMall/common/config"
	"TKMall/common/etcd"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Init("config/log.yaml", "auth")

	config.InitConfig("auth")

	serviceName := viper.GetString("server.name")
	port := viper.GetInt("server.port")
	service.SecretKey = viper.GetString("auth.secret_key")
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
		log.Infof("使用POD_IP作为服务地址: %s", serviceHost)
	} else {
		log.Infof("使用localhost作为服务地址")
	}

	err = etcd.RegisterService(client, serviceName, fmt.Sprintf("%s:%d", serviceHost, port), 10)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	auth.RegisterAuthServiceServer(server, &service.AuthServiceServer{})
	reflection.Register(server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Infof("认证服务启动成功，监听端口: %d", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
