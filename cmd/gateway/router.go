package main

import (
	"TKMall/build/proto_gen/auth"
	user "TKMall/build/proto_gen/user"
	"net/http"

	"TKMall/cmd/gateway/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func router(rpc *RPCWrapper, enforcer *casbin.Enforcer) http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())

	// 先注册所有中间件
	e.Use(middleware.AuthorizationMiddleware(enforcer)) // 先注册授权中间件
	e.Use(middleware.BlacklistMiddleware(enforcer))     // 再注册黑名单中间件

	// 然后注册路由
	e.GET("/", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":  http.StatusOK,
				"error": "Welcome To Gateway",
			},
		)
	})

	e.GET("/test_auth", rpc.Call("auth", auth.AuthServiceClient.TestGateWayMsg))

	e.POST("/login", rpc.Call("user", user.UserServiceClient.Login))
	e.POST("/register", rpc.Call("user", user.UserServiceClient.Register))

	return e
}
