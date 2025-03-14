package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/checkout"
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	"TKMall/build/proto_gen/product"
	user "TKMall/build/proto_gen/user"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
)

// 扩展 ServiceProxy 接口
type ServiceProxy interface {
	Call(ctx context.Context, service, method string, req interface{}) (interface{}, error)
	AsyncCall(ctx context.Context, service, method string, req interface{}) (<-chan Result, error)
}

// 结果包装器
type Result struct {
	Response interface{}
	Error    error
}

type GrpcProxy struct {
	mu        sync.RWMutex
	conns     map[string]*grpc.ClientConn
	endpoints map[string]string
	cache     *redis.Client
	tracer    trace.Tracer
}

// 配置熔断器
func init() {
	hystrix.ConfigureCommand("default", hystrix.CommandConfig{
		Timeout:                1000,
		MaxConcurrentRequests:  100,
		ErrorPercentThreshold:  50,
		RequestVolumeThreshold: 20,
		SleepWindow:            5000,
	})
}

// 创建新的代理实例
func NewGrpcProxy(endpoints map[string]string, redisAddr string) *GrpcProxy {
	cache := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	tracer := otel.Tracer("grpc-proxy")

	return &GrpcProxy{
		conns:     make(map[string]*grpc.ClientConn),
		endpoints: endpoints,
		cache:     cache,
		tracer:    tracer,
	}
}

// 带熔断和追踪的调用
func (p *GrpcProxy) Call(ctx context.Context, service, method string, req interface{}) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, fmt.Sprintf("%s.%s", service, method))
	defer span.End()

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("%s:%s:%v", service, method, req)
	if resp, err := p.getFromCache(ctx, cacheKey); err == nil {
		return resp, nil
	}

	var response interface{}
	err := hystrix.Do(fmt.Sprintf("%s.%s", service, method), func() error {
		var err error
		response, err = p.doCall(ctx, service, method, req)
		if err == nil {
			// 写入缓存
			p.setToCache(ctx, cacheKey, response)
		}
		return err
	}, func(err error) error {
		// 降级逻辑
		return p.fallback(ctx, service, method, req)
	})

	return response, err
}

// 异步调用
func (p *GrpcProxy) AsyncCall(ctx context.Context, service, method string, req interface{}) (<-chan Result, error) {
	resultChan := make(chan Result, 1)

	go func() {
		resp, err := p.Call(ctx, service, method, req)
		resultChan <- Result{Response: resp, Error: err}
	}()

	return resultChan, nil
}

// 缓存相关方法
func (p *GrpcProxy) getFromCache(ctx context.Context, key string) (interface{}, error) {
	val, err := p.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal([]byte(val), &result)
	return result, err
}

func (p *GrpcProxy) setToCache(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.cache.Set(ctx, key, data, time.Minute*5).Err()
}

// 降级处理
func (p *GrpcProxy) fallback(ctx context.Context, service, method string, req interface{}) error {
	// 实现降级逻辑，例如：
	// 1. 返回缓存的旧数据
	// 2. 返回默认值
	// 3. 调用备用服务
	return fmt.Errorf("service unavailable")
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
	case "product":
		return product.NewProductCatalogServiceClient(conn), nil
	case "cart":
		return cart.NewCartServiceClient(conn), nil
	case "order":
		return order.NewOrderServiceClient(conn), nil
	case "payment":
		return payment.NewPaymentServiceClient(conn), nil
	case "checkout":
		return checkout.NewCheckoutServiceClient(conn), nil
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

// 实际的 RPC 调用
func (p *GrpcProxy) doCall(ctx context.Context, service, method string, req interface{}) (interface{}, error) {
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
