package main

import (
	"TKMall/common/log"
	"fmt"
	"net"
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

	err = etcd.RegisterService(client, serviceName, fmt.Sprintf("localhost:%d", port), 10)
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
	log.Infof("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
