package service

import (
	"TKMall/build/proto_gen/cart"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CartServiceServer 购物车服务实现
type CartServiceServer struct {
	cart.UnimplementedCartServiceServer
	DB       *gorm.DB
	Redis    *redis.Client
	Node     *snowflake.Node
	Proxy    proxy.ServiceProxy
	EventBus events.EventBus
}
