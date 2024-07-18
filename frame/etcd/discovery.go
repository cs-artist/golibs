// etcd discovery
/*
服务发现:
https://qingwave.github.io/golang-distributed-system-x-etcd/

1、创建 Session， Session 中 Lease 会自动续约
2、服务注册时，在目录下创建对应的子目录，并附带 Lease
3、通过 Watch 接口监听目录变化，同步到本地
*/
package etcd

import (
	"context"
	"errors"
	"strings"
	"sync"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type DiscoveryEvent string

const (
	Put    DiscoveryEvent = "PUT"
	Delete DiscoveryEvent = "DELETE"
)

// Service
type Service struct {
	// 注册的key路径，注意放在一个目录下， e.g. /services/service_1
	Key string
	// 注册的value, e.g. ip address
	Val string

	ttl     int
	session *concurrency.Session
}

func NewService(key, val string, ttl int) (*Service, error) {
	session, err := NewSession(concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	return &Service{Key: key, Val: val, ttl: ttl, session: session}, nil
}

func (s *Service) Register(ctx context.Context) error {
	client := s.session.Client()
	_, err := client.Put(ctx, s.Key, s.Val, clientv3.WithLease(s.session.Lease()))
	return err
}

func (s *Service) UnRegister(ctx context.Context) error {
	client := s.session.Client()
	_, err := client.Delete(ctx, s.Key)
	return err
}

func (s *Service) Close() error {
	return s.session.Close()
}

// Discovery
type Discovery struct {
	// Key prefix for services to register with
	Prefix string
	// TTL is the etcd session's TTL in seconds
	TTL int
	// Callbacks are callbacks that are triggered during certain lifecycle events
	Callbacks DiscoveryCallbacks
	// Services list
	services map[string]string
	locker   sync.RWMutex

	session *concurrency.Session
}

type DiscoveryCallbacks struct {
	OnStartedDiscovering func()
	OnStoppedDiscovering func()
	OnServiceChanged     func(event DiscoveryEvent, service *Service)
}

func NewDiscovery(prefix string, ttl int, cbs DiscoveryCallbacks) (*Discovery, error) {
	session, err := NewSession(concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	prefix = strings.TrimSuffix(prefix, "/") + "/"
	return &Discovery{Prefix: prefix, TTL: ttl, Callbacks: cbs, session: session}, nil
}

// Watch 监听目录变化
func (d *Discovery) Watch(ctx context.Context) error {
	client := d.session.Client()
	// get keys with matching prefix
	rsp, err := client.Get(ctx, d.Prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// 初始化service
	services := make(map[string]string)
	for _, kv := range rsp.Kvs {
		services[string(kv.Key)] = string(kv.Value)
	}
	d.setServices(services)

	d.Callbacks.OnStartedDiscovering()

	// watch
	ch := client.Watch(ctx, d.Prefix, clientv3.WithPrefix())
	defer d.Callbacks.OnStoppedDiscovering()

	for {
		select {
		case <-d.session.Done():
			// closes when the lease is orphaned, expires, or is otherwise no longer being refreshed
			return nil
		case rsp, ok := <-ch:
			// The channel closes when the context is canceled or the underlying watcher
			// is otherwise disrupted.
			if !ok {
				return errors.New("watch closed")
			}
			if err := rsp.Err(); err != nil {
				return err
			}
			for _, event := range rsp.Events {
				key, val := string(event.Kv.Key), string(event.Kv.Value)
				var triggerEvent DiscoveryEvent
				switch event.Type {
				case mvccpb.PUT:
					d.addService(key, val)
					triggerEvent = Put
				case mvccpb.DELETE:
					triggerEvent = Delete
					d.delService(key)
				}
				d.Callbacks.OnServiceChanged(triggerEvent, &Service{Key: key, Val: val})
			}
		}
	}
}

func (d *Discovery) Close() error {
	return d.session.Close()
}

// DelServices 删除service列表
func (d *Discovery) DelServices(ctx context.Context) error {
	client := d.session.Client()
	_, err := client.Delete(ctx, d.Prefix, clientv3.WithPrefix())
	return err
}

// GetServices 获取service列表
func (d *Discovery) GetServices() []*Service {
	d.locker.RLocker()
	defer d.locker.RUnlock()
	services := make([]*Service, 0, len(d.services))
	for key, val := range d.services {
		services = append(services, &Service{
			Key: key,
			Val: val,
		})
	}
	return services
}

func (d *Discovery) setServices(services map[string]string) {
	d.locker.Lock()
	defer d.locker.Unlock()
	d.services = services
}

func (d *Discovery) addService(key, val string) {
	d.locker.Lock()
	defer d.locker.Unlock()
	d.services[key] = val
}

func (d *Discovery) delService(key string) {
	d.locker.Lock()
	defer d.locker.Unlock()
	delete(d.services, key)
}
