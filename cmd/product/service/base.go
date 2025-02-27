package service

import (
	"TKMall/build/proto_gen/product"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductCatalogServiceServer struct {
	product.UnimplementedProductCatalogServiceServer
	DB       *gorm.DB
	Redis    *redis.Client
	Node     *snowflake.Node
	Proxy    proxy.ServiceProxy
	EventBus events.EventBus
}

// type ProductService struct {
// }
