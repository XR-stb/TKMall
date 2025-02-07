package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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
	log.Printf("租约ID: %d", grantResp.ID)
	_, err = etcdClient.Put(context.Background(), serviceName, serviceAddr, clientv3.WithLease(grantResp.ID))
	if err != nil {
		return err
	}
	log.Printf("服务 %s 已注册到 ETCD，租约ID: %d", serviceName, grantResp.ID)
	go func() {
		for {
			_, err := lease.KeepAliveOnce(context.Background(), grantResp.ID)
			if err != nil {
				log.Printf("Failed to keep alive: %v", err)
				break
			}
			time.Sleep(time.Duration(ttl/2) * time.Second)
		}
	}()

	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置: %+v", config)
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer etcdClient.Close()

	serviceName := config.Server.Name
	serviceAddr := fmt.Sprintf("localhost:%d", config.Server.Port)
	log.Printf("服务 %s 将注册到 ETCD，地址为 %s", serviceName, serviceAddr)
	if err := registerService(etcdClient, serviceName, serviceAddr, 10); err != nil {
		log.Fatalf("服务注册失败: %v", err)
	}
	log.Printf("服务 %s 已注册到 ETCD，地址为 %s", serviceName, serviceAddr)

	server01 := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      router(etcdClient, config),
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
