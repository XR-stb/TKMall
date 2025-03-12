package service

import (
	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/checkout"
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CheckoutServiceServer 结账服务实现
type CheckoutServiceServer struct {
	checkout.UnimplementedCheckoutServiceServer
	DB             *gorm.DB
	Redis          *redis.Client
	Node           *snowflake.Node
	Proxy          proxy.ServiceProxy
	EventBus       events.EventBus
	OrderService   order.OrderServiceClient     // 订单服务客户端
	PaymentService payment.PaymentServiceClient // 支付服务客户端
	CartService    cart.CartServiceClient       // 购物车服务客户端
}

// 将proto Address转换为order proto中的Address
func convertProtoAddress(addr *checkout.Address) *order.Address {
	if addr == nil {
		return nil
	}
	return &order.Address{
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		ZipCode:       int32(0), // 需要字符串转换为int32
	}
}

// 计算购物车总金额（实际应用中可能更复杂）
func calculateCartTotal(items []*cart.CartItem) float32 {
	var total float32
	for _, item := range items {
		// 简化版，实际应该查询商品价格
		total += float32(item.Quantity) * 100.0 // 假设每个商品100元
	}
	return total
}
