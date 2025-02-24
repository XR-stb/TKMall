package proxy

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"TKMall/build/proto_gen/auth"
	user "TKMall/build/proto_gen/user"

	"google.golang.org/grpc"
)

type ServiceProxy interface {
	Call(ctx context.Context, service, method string, req interface{}) (interface{}, error)
}

type GrpcProxy struct {
	mu        sync.RWMutex
	conns     map[string]*grpc.ClientConn
	endpoints map[string]string
}

func NewGrpcProxy(endpoints map[string]string) *GrpcProxy {
	return &GrpcProxy{
		conns:     make(map[string]*grpc.ClientConn),
		endpoints: endpoints,
	}
}

func (p *GrpcProxy) Call(ctx context.Context, service, method string, req interface{}) (interface{}, error) {
	conn, err := p.getConnection(service)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %v", err)
	}

	// 根据服务名和方法名获取对应的客户端和方法
	client, err := p.getServiceClient(conn, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %v", err)
	}

	// 通过反射调用方法
	result, err := p.invokeMethod(ctx, client, method, req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke method: %v", err)
	}

	return result, nil
}

func (p *GrpcProxy) getConnection(service string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, exists := p.conns[service]
	p.mu.RUnlock()

	if exists {
		return conn, nil
	}

	endpoint, ok := p.endpoints[service]
	if !ok {
		return nil, fmt.Errorf("service endpoint not found: %s", service)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check
	if conn, exists = p.conns[service]; exists {
		return conn, nil
	}

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	p.conns[service] = conn
	return conn, nil
}

func (p *GrpcProxy) getServiceClient(conn *grpc.ClientConn, service string) (interface{}, error) {
	// 根据服务名获取对应的 NewXXXClient 函数
	switch service {
	case "auth":
		return auth.NewAuthServiceClient(conn), nil
	case "user":
		return user.NewUserServiceClient(conn), nil
	default:
		return nil, fmt.Errorf("unknown service: %s. please in here implement or fix your code.", service)
	}
}

func (p *GrpcProxy) invokeMethod(ctx context.Context, client interface{}, methodName string, req interface{}) (interface{}, error) {
	// 获取客户端的反射值
	val := reflect.ValueOf(client)

	// 获取方法
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	// 调用方法
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	}

	results := method.Call(args)

	// gRPC 方法总是返回 (response, error)
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected number of return values")
	}

	// 检查错误
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}
