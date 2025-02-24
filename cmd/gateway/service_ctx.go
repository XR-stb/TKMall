package main

import (
	"TKMall/build/proto_gen/auth"
	user "TKMall/build/proto_gen/user"
	"TKMall/common/log"
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc"
)

type ServiceContext struct {
	clients     map[string]interface{} // 使用map存储所有客户端
	connections map[string]*grpc.ClientConn
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
		"user": {cfg.Services.UserService, func(conn grpc.ClientConnInterface) interface{} {
			return user.NewUserServiceClient(conn)
		}},
		"auth": {cfg.Services.AuthService, func(conn grpc.ClientConnInterface) interface{} {
			return auth.NewAuthServiceClient(conn)
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
