package nx_etcd

import (
	"context"
	"errors"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock/pkg"
	"go.etcd.io/etcd/v3/clientv3"
)

type EtcdV3 struct {
	*clientv3.Client

	config clientv3.Config

	logger log.Logger
}

type OptionV3 func(c *EtcdV3)

// 分布式锁etcdv3 解决方案
func NewEtcdV3(options ...OptionV3) pkg.NxlockSolution {
	etcdv3 := &EtcdV3{}

	for _, option := range options {
		option(etcdv3)
	}

	if etcdv3.logger == nil {
		etcdv3.logger = log.NewLogger()
	}

	cli, err := clientv3.New(etcdv3.config)
	if err != nil {
		etcdv3.logger.Fatalf(err.Error())
	}

	etcdv3.Client = cli

	return etcdv3
}

func WithLogger(logger log.Logger) OptionV3 {
	return func(e *EtcdV3) {
		e.logger = logger
	}
}

func WithConfig(config clientv3.Config) OptionV3 {
	return func(e *EtcdV3) {
		e.config = config
	}
}

func (e3 *EtcdV3) Lock(ctx context.Context, key, val string, ttl int64) error {
	grResp, err := e3.Client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	err = e3.keepAlive(ctx, grResp.ID)
	if err != nil {
		return err
	}

	resp := &clientv3.TxnResponse{}
	// 使用事物获取锁
	resp, err = e3.Client.Txn(ctx).If(
		clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
	).Then(
		clientv3.OpPut(key, val, clientv3.WithLease(grResp.ID)),
	).Commit()

	// 锁已经存在，或者网络原因失败，删除合约
	if err != nil || !resp.Succeeded {
		relErr := e3.Release(ctx, key)
		if relErr != nil {
			e3.logger.Errorc(ctx, relErr.Error())
		}

		return errors.New(pkg.ErrEtcdV3LockFailure)
	}

	return nil
}

func (e3 *EtcdV3) Release(ctx context.Context, key string) error {
	resp, err := e3.Client.Delete(ctx, key)
	e3.logger.Debugc(ctx, "Release %s : %+v", key, resp)
	return err
}

// 续租
func (e3 *EtcdV3) keepAlive(ctx context.Context, leaseID clientv3.LeaseID) error {
	kaResp, err := e3.Client.KeepAlive(ctx, leaseID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case resp, ok := <-kaResp:
				if ok {
					if resp != nil {
						e3.logger.Errorc(ctx, "keepAlive %d : %+v", leaseID, resp)
					}
				} else {
					e3.logger.Errorc(ctx, "LeaseKeepAlive %d is close.", leaseID)
					return
				}
			}
		}
	}()

	return nil
}

func (e3 *EtcdV3) Close() error {
	return e3.Client.Close()
}
