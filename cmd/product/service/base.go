package service

import (
	"TKMall/build/proto_gen/product"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type ProductCatalogServiceServer struct {
	product.UnimplementedProductCatalogServiceServer
	DB       *gorm.DB
	Node     *snowflake.Node
	Proxy    proxy.ServiceProxy
	EventBus events.EventBus
}

// type ProductService struct {
// }
