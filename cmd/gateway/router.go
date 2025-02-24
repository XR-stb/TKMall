package main

import (
	"TKMall/build/proto_gen/auth"
	user "TKMall/build/proto_gen/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func router(rpc *RPCWrapper) http.Handler {
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

	e.GET("/test_auth", rpc.Call("auth", auth.AuthServiceClient.TestGateWayMsg))

	e.POST("/login", rpc.Call("user", user.UserServiceClient.Login))
	e.POST("/register", rpc.Call("user", user.UserServiceClient.Register))

	return e
}
