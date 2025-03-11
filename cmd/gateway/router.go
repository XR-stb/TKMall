package main

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/cart"
	product "TKMall/build/proto_gen/product"
	user "TKMall/build/proto_gen/user"
	"TKMall/common/log"
	"net/http"
	"time"

	"TKMall/cmd/gateway/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func router(rpc *RPCWrapper, enforcer *casbin.Enforcer) http.Handler {
	// 加载白名单配置
	if err := middleware.LoadWhitelistConfig(); err != nil {
		log.Fatalf("初始化白名单配置失败: %v", err)
	}

	e := gin.New()
	e.Use(gin.Recovery())

	// 先注册黑名单中间件
	e.Use(middleware.BlacklistMiddleware(enforcer))     // 先注册黑名单中间件
	e.Use(middleware.AuthorizationMiddleware(enforcer)) // 再注册授权中间件

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

	// 添加商品服务路由
	productGroup := e.Group("/products")
	{
		productGroup.GET("", rpc.Call("product", product.ProductCatalogServiceClient.ListProducts))
		productGroup.GET("/get", middleware.CacheMiddleware(5*time.Minute), rpc.Call("product", product.ProductCatalogServiceClient.GetProduct))
		productGroup.POST("/search", rpc.Call("product", product.ProductCatalogServiceClient.SearchProducts))
	}

	// 添加购物车服务路由
	cartGroup := e.Group("/cart")
	{
		cartGroup.POST("/add", rpc.Call("cart", cart.CartServiceClient.AddItem))
		cartGroup.GET("/get", rpc.Call("cart", cart.CartServiceClient.GetCart))
		cartGroup.DELETE("/empty", rpc.Call("cart", cart.CartServiceClient.EmptyCart))
	}

	return e
}
