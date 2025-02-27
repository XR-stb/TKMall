package main

import (
	"TKMall/common/log"
	"context"
	"fmt"
	"net/http"
	"time"

	"TKMall/cmd/gateway/middleware"

	"github.com/casbin/casbin/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func registerService(etcdClient *clientv3.Client, serviceName, serviceAddr string, ttl int) error {
	lease := clientv3.NewLease(etcdClient)
	grantResp, err := lease.Grant(context.Background(), int64(ttl))
	if err != nil {
		return err
	}
	log.Infof("租约ID: %d", grantResp.ID)
	_, err = etcdClient.Put(context.Background(), serviceName, serviceAddr, clientv3.WithLease(grantResp.ID))
	if err != nil {
		return err
	}
	log.Infof("服务 %s 已注册到 ETCD，租约ID: %d", serviceName, grantResp.ID)
	go func() {
		for {
			_, err := lease.KeepAliveOnce(context.Background(), grantResp.ID)
			if err != nil {
				log.Infof("Failed to keep alive: %v", err)
				break
			}
			time.Sleep(time.Duration(ttl/2) * time.Second)
		}
	}()

	return nil
}

func main() {
	log.Init("config/log.yaml", "gateway")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化 etcd 客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Etcd.Endpoints,
		DialTimeout: time.Duration(config.Etcd.DialTimeout) * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	defer etcdClient.Close()

	// 初始化 Casbin
	// TODO: 后续权限记录应该写入db，不再读取csv，github.com/casbin/xorm-adapter 已做支持
	e, err := casbin.NewEnforcer("config/casbin/model.conf", "config/casbin/policy.csv")
	if err != nil {
		log.Fatalf("Failed to initialize Casbin: %v", err)
	}

	// 创建 Casbin 适配器
	adapter := middleware.NewCasbinAdapter(etcdClient, e)

	// 初始化时加载策略到 etcd（如果是第一次运行）
	if err := adapter.SavePolicy(); err != nil {
		log.Errorf("Failed to save initial policy to etcd: %v", err)
	}

	// 初始化 Enforcer
	middleware.InitEnforcer(e)

	serviceCtx := NewServiceContext(config)
	rpcWrapper := NewRPCWrapper(serviceCtx)

	// 使用现有的 router 函数
	r := router(rpcWrapper, e)

	server01 := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
