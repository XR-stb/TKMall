package main

import (
	user "TKMall/build/proto_gen/user"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func handleGRPCRequest(c *gin.Context, etcdClient *clientv3.Client, serviceName string, grpcFunc func(user.UserServiceClient, context.Context) (interface{}, error)) {
	// 发现 user 服务
	serviceAddr, err := discoverService(etcdClient, serviceName)
	if err != nil || serviceAddr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": "user service not found",
		})
		return
	}

	// 获取 gRPC 客户端
	userClient, conn, err := getUserClient(serviceAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()

	// 调用 gRPC 服务
	resp, err := grpcFunc(userClient, context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func discoverService(client *clientv3.Client, serviceName string) (string, error) {
	// 若Etcd服务器未启动，此处会阻塞
	resp, err := client.Get(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

func getUserClient(serviceAddr string) (user.UserServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(serviceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	client := user.NewUserServiceClient(conn)
	return client, conn, nil
}
