// etcd election
/*
  Simple use:

  e := concurrency.NewElection(my_session, "/leader/")
  go e.Campaign(ctx, id)

  for {
    leader, _ := e.Leader(ctx)
    if leader.Kvs[0].Value == id {
      print("I am the leader")
    } else {
      print("I am the replica")
    }
    time.Sleep(30*time.Second)
  }
*/
package etcd

import (
	"context"

	"go.etcd.io/etcd/client/v3/concurrency"
)

// Implement Master election with etcd in a master/slave application cluster
// https://github.com/zhangwuh/etcd-master-election
/*
ch, _ := Campaign(ctx, "/election-path", "app", leaseTtl)
for {
	select {
	case isMaster := <- ch:
		if isMaster {
			fmt.Println("i'm the master now")
			//do master's work
		}
	}
}
*/
func Campaign(ctx context.Context, prefix string, val string, ttl int) (<-chan bool, error) {
	leader := make(chan bool)
	go func() {
		defer close(leader) // 退出时关闭
		for {
			// step1: 创建租约
			// session中keepAlive会一直续租，如果续租失败，session.Done()会收到退出信号
			session, err := NewSession(concurrency.WithTTL(ttl))
			if err != nil {
				continue
			}

			// step2: 开始竞选
			election := concurrency.NewElection(session, prefix)
			// leader节点返回成功，其他节点阻塞等待
			if err = election.Campaign(ctx, val); err != nil {
				session.Close()
				// 竞选取消
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				// 竞选失败，继续
				continue
			}

			// 竞选成功
			leader <- true

			// step3: 阻塞等待
			select {
			case <-session.Done():
				// keepAlive失败，继续下次选举
				session.Close()
				leader <- false
			case <-ctx.Done():
				// 外部取消，退出
				session.Close()
				return
			}
		}
	}()
	return leader, nil
}
