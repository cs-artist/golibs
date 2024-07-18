package etcd

import (
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var defaultClient *clientv3.Client

// InitDefaultClient 初始化
func InitDefaultClient(cfg *clientv3.Config) error {
	client, err := clientv3.New(*cfg)
	if err != nil {
		return err
	}
	defaultClient = client
	return nil
}

// GetDefaultClient 获取默认client
func GetDefaultClient() *clientv3.Client {
	return defaultClient
}

// NewSession 创建一个lease，默认是60s TTL，并会调用KeepAlive，
// 永久为这个lease自动续约（2/3生命周期的时候执行续约操作）
// e.g. NewSession(concurrency.WithTTL(5))
func NewSession(opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	client := GetDefaultClient()
	if client == nil {
		return nil, errors.New("client not init")
	}
	return concurrency.NewSession(client, opts...)
}
