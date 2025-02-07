package main

import (
	user "TKMall/build/proto_gen/user"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func router(etcdClient *clientv3.Client, config *Config) http.Handler {
	// 创建服务上下文
	serviceCtx := NewServiceContext(config)
	_ = serviceCtx

	e := gin.New()
	e.Use(gin.Recovery())

	e.GET("/", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":  http.StatusOK,
				"error": "Welcome To Gateway",
			},
		)
	})

	e.POST("/login", func(c *gin.Context) {
		handleGRPCRequest(c, etcdClient, "user-service", func(userClient user.UserServiceClient, ctx context.Context) (interface{}, error) {
			email := c.Query("email")
			password := c.Query("password")
			req := &user.LoginReq{
				Email:    email,
				Password: password,
			}
			return userClient.Login(ctx, req)
		})
	})

	e.POST("/register", func(c *gin.Context) {
		handleGRPCRequest(c, etcdClient, "user-service", func(userClient user.UserServiceClient, ctx context.Context) (interface{}, error) {
			email := c.Query("email")
			password := c.Query("password")
			req := &user.RegisterReq{
				Email:    email,
				Password: password,
			}
			return userClient.Register(ctx, req)
		})
	})

	return e
}
