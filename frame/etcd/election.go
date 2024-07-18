// etcd election
/*
etcd选举逻辑由Election.Campaign实现，逻辑与 etcd 锁中的实现非常相似：
1、开启事务，首先判断当前服务 Key 是否存在
2、不存在，通过 Put 设置对应值
3、存在获得当前目录最小 Revision 的值，即当前主节点
4、通过 waitDeletes，直到当前进程的 Revision

https://qingwave.github.io/golang-distributed-system-x-etcd/
*/
package etcd

import (
	"context"

	"go.etcd.io/etcd/client/v3/concurrency"
)

type Election struct {
	// Key prefix for election
	Prefix string
	// Proposal is the value as eligible for the election on the prefix key
	Proposal string
	// TTL is the duration that non-leader candidates will wait to force acquire leadership
	TTL int
	// Callbacks are callbacks that are triggered during certain lifecycle events of the LeaderElector
	Callbacks ElectionCallbacks
	session   *concurrency.Session
	election  *concurrency.Election
}

type ElectionCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(context.Context)
	// OnStoppedLeading is called when a LeaderElector client stops leading
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader func(proposal string)
}

func NewElection(prefix string, proposal string, ttl int, cbs ElectionCallbacks) (*Election, error) {
	session, err := NewSession(concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	election := concurrency.NewElection(session, prefix)
	return &Election{
		Prefix:    prefix,
		Proposal:  proposal,
		TTL:       ttl,
		Callbacks: cbs,
		session:   session,
		election:  election,
	}, nil
}

func (e *Election) Campaign(ctx context.Context) error {
	go e.observe(ctx)

	err := e.election.Campaign(ctx, e.Proposal)
	if err != nil {
		return err
	}

	e.Callbacks.OnStartedLeading(ctx)
	return nil
}

func (e *Election) observe(ctx context.Context) {
	ch := e.election.Observe(ctx)
	for {
		select {
		case <-e.session.Done():
			// closes when the lease is orphaned, expires, or is otherwise no longer being refreshed
			e.Callbacks.OnStoppedLeading()
			return
		case rsp, ok := <-ch:
			// The channel closes when the context is canceled or the underlying watcher
			// is otherwise disrupted.
			if !ok {
				return
			}
			for _, kv := range rsp.Kvs {
				if leader := string(kv.Value); leader != e.Proposal {
					e.Callbacks.OnNewLeader(leader)
				}
			}
		}
	}
}

func (e *Election) Close() error {
	return e.session.Close()
}
