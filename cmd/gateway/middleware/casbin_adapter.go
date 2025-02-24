package middleware

import (
	"context"
	"encoding/json"
	"fmt"

	"TKMall/common/log"

	"github.com/casbin/casbin/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	PolicyKey = "/casbin/policy" // etcd 中存储策略的键
)

type CasbinAdapter struct {
	client    *clientv3.Client
	enforcer  *casbin.Enforcer
	watchChan clientv3.WatchChan
}

// 创建新的适配器
func NewCasbinAdapter(client *clientv3.Client, enforcer *casbin.Enforcer) *CasbinAdapter {
	adapter := &CasbinAdapter{
		client:   client,
		enforcer: enforcer,
	}

	// 开始监听策略变更
	adapter.watchPolicyChanges()
	return adapter
}

// 监听策略变更
func (a *CasbinAdapter) watchPolicyChanges() {
	a.watchChan = a.client.Watch(context.Background(), PolicyKey)
	go func() {
		for watchResp := range a.watchChan {
			for _, event := range watchResp.Events {
				log.Infof("检测到策略变更: %s", event.Type)
				if err := a.loadPolicyFromEtcd(); err != nil {
					log.Errorf("加载策略失败: %v", err)
				}
			}
		}
	}()
}

// 从 etcd 加载策略
func (a *CasbinAdapter) loadPolicyFromEtcd() error {
	resp, err := a.client.Get(context.Background(), PolicyKey)
	if err != nil {
		return fmt.Errorf("从 etcd 获取策略失败: %v", err)
	}

	if len(resp.Kvs) == 0 {
		return nil
	}

	var policies [][]string
	if err := json.Unmarshal(resp.Kvs[0].Value, &policies); err != nil {
		return fmt.Errorf("解析策略失败: %v", err)
	}

	// 清除现有策略并加载新策略
	a.enforcer.ClearPolicy()
	for _, policy := range policies {
		if len(policy) == 3 {
			// 将 []string 转换为 []interface{}
			params := make([]interface{}, len(policy))
			for i, v := range policy {
				params[i] = v
			}
			_, err = a.enforcer.AddPolicy(params...)
		} else if len(policy) == 2 {
			params := make([]interface{}, len(policy))
			for i, v := range policy {
				params[i] = v
			}
			_, err = a.enforcer.AddGroupingPolicy(params...)
		}
		if err != nil {
			log.Errorf("添加策略失败: %v", err)
		}
	}

	return nil
}

// 保存策略到 etcd
func (a *CasbinAdapter) SavePolicy() error {
	// 获取所有策略
	policies, err := a.enforcer.GetPolicy()
	if err != nil {
		return err
	}

	groupingPolicies, err := a.enforcer.GetGroupingPolicy()
	if err != nil {
		return err
	}

	// 合并策略
	allPolicies := append(policies, groupingPolicies...)

	// 序列化策略
	data, err := json.Marshal(allPolicies)
	if err != nil {
		return fmt.Errorf("序列化策略失败: %v", err)
	}

	// 保存到 etcd
	_, err = a.client.Put(context.Background(), PolicyKey, string(data))
	if err != nil {
		return fmt.Errorf("保存策略到 etcd 失败: %v", err)
	}

	log.Info("策略已成功保存到 etcd")
	return nil
}
