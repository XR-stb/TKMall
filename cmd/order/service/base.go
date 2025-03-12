package service

import (
	"TKMall/build/proto_gen/order"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// OrderServiceServer 订单服务实现
type OrderServiceServer struct {
	order.UnimplementedOrderServiceServer
	DB       *gorm.DB
	Redis    *redis.Client
	Node     *snowflake.Node
	Proxy    proxy.ServiceProxy
	EventBus events.EventBus
}

// Address结构体
type Address struct {
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	ZipCode       int32  `json:"zip_code"`
}

// 将proto中的Address转换为model中的Address
func convertToModelAddress(addr *order.Address) Address {
	if addr == nil {
		return Address{}
	}
	return Address{
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		ZipCode:       addr.ZipCode,
	}
}

// 将model中的Address转换为proto中的Address
func convertToProtoAddress(addr Address) *order.Address {
	return &order.Address{
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		ZipCode:       addr.ZipCode,
	}
}
