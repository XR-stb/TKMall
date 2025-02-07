package main

import (
	user "TKMall/build/proto_gen/user"
	"log"

	"google.golang.org/grpc"
)

type ServiceContext struct {
	UserClient user.UserServiceClient
}

func NewServiceContext(cfg *Config) *ServiceContext {
	maxSize := 20 * 1024 * 1024
	diaOpt := grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxSize))

	conn, err := grpc.Dial(
		cfg.Services.UserService,
		grpc.WithInsecure(),
		diaOpt,
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	return &ServiceContext{
		UserClient: user.NewUserServiceClient(conn),
	}
}
