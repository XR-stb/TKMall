package etcd

import (
	"context"
	"fmt"
	"os"

	"TKMall/common/log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func RegisterService(client *clientv3.Client, serviceName, serviceAddr string, ttl int64) error {
	// 在Kubernetes环境中，使用不同的服务注册方式
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// 从环境变量获取POD_IP
		podIP := os.Getenv("POD_IP")
		if podIP != "" {
			// 使用POD_IP替换localhost
			if serviceAddr == fmt.Sprintf("localhost:%d", 50051) {
				port := 50051 // 默认端口，也可从环境变量获取
				serviceAddr = fmt.Sprintf("%s:%d", podIP, port)
				log.Infof("在Kubernetes环境中注册服务，地址: %s", serviceAddr)
			}
		}
	}

	leaseResp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	// 在ETCD中注册服务, 使用 "/services/服务名" 作为key
	serviceKey := fmt.Sprintf("/services/%s", serviceName)
	_, err = client.Put(context.Background(), serviceKey, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	ch, err := client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-ch
		}
	}()

	log.Infof("服务 %s 已在 %s 注册到ETCD", serviceName, serviceAddr)
	return nil
}
