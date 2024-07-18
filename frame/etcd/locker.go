// etcd locker
/*
Etcd 分布式锁的步骤:
https://pandaychen.github.io/2019/10/24/ETCD-DISTRIBUTED-LOCK/

1、假设分布式锁的 Name 为 /root/lockname，用来控制某个共享资源，concurrency 会自动将其转换为目录形式：/root/lockname/

2、客户端 A 连接 Etcd，创建一个租约 Leaseid_A，并设置 TTL（以业务逻辑来定 TTL 的时间）, 以 /root/lockname 为前缀创建全局唯一的 Key，
该 Key 的组织形式为 /root/lockname/{leaseid_A}，同时调用 TXN 条件事务：
（1）判断条件：比较/root/lockname/{leaseid_A}的 CreateRevision 是否为 0
（2）等于 0 表示目前不存在该 key，Then(put, getOwner)，客户端 A 将此 Key 绑定租约写入 Etcd，同时调用 TXN 事务查询写入的情况和具有相
同前缀 / root/lockname / 的 CreateRevision的排序情况；后续直接返回此key相关信息
（3）不等于 0 表示对应的 key 已经创建了，Else(get, getOwner)，以前缀 /root/lockname/ 读取 keyValue 列表（keyValue 中带有 key
对应的 CreateRevision），判断自己 key 的 CreateRevision是否为当前列表中最小的，如果是则认为获得锁；否则阻塞监听列表中前一个
CreateRevision比自己小的 key 的删除事件，一旦监听到删除事件或者因租约失效而删除的事件，则自己获得锁

3、执行业务逻辑，操作共享资源

4、释放分布式锁，现网的程序逻辑需要实现在正常和异常条件下的释放锁的策略，如捕获 SIGTERM 后执行 Unlock，或者异常退出时，有完善的监控和及时
删除 Etcd 中的 Key 的异步机制，避免出现死锁现象

5、当客户端持有锁期间，其它客户端只能等待，为了避免等待期间租约失效，客户端需创建一个定时任务进行续约续期。如果持有锁期间客户端崩溃，心跳停止，
Key 将因租约到期而被删除，从而锁释放，避免死锁

注意事项：
1、Etcd client 有 V2 和 V3 版本，数据是不互通的，所以勿用 V3 的 API 去操作 V2 版本 API 写入的数据，反之亦然
2、V3版本 Concurrency包 封装了创建租约，自动续约等操作，提供了Lock、Unlock等接口
https://pkg.go.dev/github.com/coreos/etcd/clientv3/concurrency
*/
package etcd

import (
	"context"

	"go.etcd.io/etcd/client/v3/concurrency"
)

type Locker struct {
	Prefix  string
	TTL     int
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func NewLocker(prefix string, ttl int) (*Locker, error) {
	session, err := NewSession(concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}
	mutex := concurrency.NewMutex(session, prefix)
	return &Locker{Prefix: prefix, TTL: ttl, session: session, mutex: mutex}, nil
}

func (l *Locker) Destory() error {
	return l.session.Close()
}

func (l *Locker) Trylock(ctx context.Context) (bool, error) {
	err := l.mutex.TryLock(ctx)
	if err == nil {
		return true, nil
	}
	if err == concurrency.ErrLocked {
		return false, nil
	}
	return false, err
}

func (l *Locker) Lock(ctx context.Context) error {
	return l.mutex.Lock(ctx)
}

func (l *Locker) Unlock(ctx context.Context) error {
	return l.mutex.Unlock(ctx)
}
