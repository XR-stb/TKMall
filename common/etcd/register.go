package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func RegisterService(client *clientv3.Client, serviceName, serviceAddr string, ttl int64) error {
	leaseResp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	_, err = client.Put(context.Background(), serviceName, serviceAddr, clientv3.WithLease(leaseResp.ID))
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

	return nil
}
