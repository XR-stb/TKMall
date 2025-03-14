package main

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/checkout"
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	product "TKMall/build/proto_gen/product"
	user "TKMall/build/proto_gen/user"
	"TKMall/common/log"
	"context"
	"fmt"
	"os"
	"reflect"

	"google.golang.org/grpc"
)

type ServiceContext struct {
	clients     map[string]interface{} // 使用map存储所有客户端
	connections map[string]*grpc.ClientConn
}

// 从环境变量获取服务地址，如果存在则使用环境变量，否则使用配置
func getServiceAddr(envName, defaultAddr string) string {
	if addr := os.Getenv(envName); addr != "" {
		log.Infof("使用环境变量地址 %s: %s", envName, addr)
		return addr
	}
	log.Infof("使用默认配置地址: %s", defaultAddr)
	return defaultAddr
}

// 初始化服务连接
func NewServiceContext(cfg *Config) *ServiceContext {
	sc := &ServiceContext{
		clients:     make(map[string]interface{}),
		connections: make(map[string]*grpc.ClientConn),
	}

	// 统一配置项
	services := map[string]struct {
		address  string
		clientFn func(conn grpc.ClientConnInterface) interface{}
	}{
		"user": {getServiceAddr("USER_SERVICE_ADDR", cfg.Services.UserService), func(conn grpc.ClientConnInterface) interface{} {
			return user.NewUserServiceClient(conn)
		}},
		"auth": {getServiceAddr("AUTH_SERVICE_ADDR", cfg.Services.AuthService), func(conn grpc.ClientConnInterface) interface{} {
			return auth.NewAuthServiceClient(conn)
		}},
		"product": {getServiceAddr("PRODUCT_SERVICE_ADDR", cfg.Services.ProductService), func(conn grpc.ClientConnInterface) interface{} {
			return product.NewProductCatalogServiceClient(conn)
		}},
		"order": {getServiceAddr("ORDER_SERVICE_ADDR", cfg.Services.OrderService), func(conn grpc.ClientConnInterface) interface{} {
			return order.NewOrderServiceClient(conn)
		}},
		"payment": {getServiceAddr("PAYMENT_SERVICE_ADDR", cfg.Services.PaymentService), func(conn grpc.ClientConnInterface) interface{} {
			return payment.NewPaymentServiceClient(conn)
		}},
		"checkout": {getServiceAddr("CHECKOUT_SERVICE_ADDR", cfg.Services.CheckoutService), func(conn grpc.ClientConnInterface) interface{} {
			return checkout.NewCheckoutServiceClient(conn)
		}},
		"cart": {getServiceAddr("CART_SERVICE_ADDR", cfg.Services.CartService), func(conn grpc.ClientConnInterface) interface{} {
			return cart.NewCartServiceClient(conn)
		}},
	}

	maxSize := 20 * 1024 * 1024
	diaOpt := grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxSize))

	for name, service := range services {
		conn, err := createGRPCConnection(service.address, diaOpt)
		if err != nil {
			log.Fatalf("[%s] 服务连接失败: %v", name, err)
		}

		sc.connections[name] = conn
		sc.clients[name] = service.clientFn(conn)
	}

	return sc
}

// 通用连接创建方法
func createGRPCConnection(address string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	baseOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}
	return grpc.DialContext(context.Background(), address, append(baseOpts, opts...)...)
}

// 通用获取客户端方法（带类型断言）
func (s *ServiceContext) GetClient(name string, client interface{}) error {
	if c, ok := s.clients[name]; ok {
		// 使用反射进行类型安全赋值
		val := reflect.ValueOf(client)
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("client must be a pointer")
		}
		val.Elem().Set(reflect.ValueOf(c))
		return nil
	}
	return fmt.Errorf("client %s not found", name)
}

// 统一关闭所有连接
func (s *ServiceContext) Close() {
	for name, conn := range s.connections {
		if err := conn.Close(); err != nil {
			log.Infof("[%s] 连接关闭失败: %v", name, err)
		}
	}
}
