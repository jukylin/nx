package nxlock

import (
	"context"
	"errors"
	"sync"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock/pkg"
	"go.etcd.io/etcd/v3/etcdserver/api/v3rpc/rpctypes"
	"fmt"
)

type Option func(*Nxlock)

type Nxlock struct {
	logger log.Logger

	// 分布式锁解决方案
	Solution pkg.NxlockSolution

	// 重试次数
	retryTime int

	// 当前进程已获得锁
	holdLock *sync.Map
}

func NewNxlock(options ...Option) *Nxlock {
	nl := &Nxlock{}

	for _, option := range options {
		option(nl)
	}

	if nl.retryTime == 0 {
		nl.retryTime = 3
	}

	nl.holdLock = &sync.Map{}

	return nl
}

func WithLogger(logger log.Logger) Option {
	return func(nl *Nxlock) {
		nl.logger = logger
	}
}

func WithSolution(solution pkg.NxlockSolution) Option {
	return func(nl *Nxlock) {
		nl.Solution = solution
	}
}

func WithRetryLock(retryTime int) Option {
	return func(nl *Nxlock) {
		nl.retryTime = retryTime
	}
}

func (nl *Nxlock) Lock(ctx context.Context, key string, ttl int64) error {
	var err error

	// 避免资源挣抢
	_, loaded := nl.holdLock.LoadOrStore(key, true)
	if loaded {
		return fmt.Errorf(pkg.ErrAlreadyAcquiredLock, key)
	}

	for i := 0; i < nl.retryTime; i++ {
		err = nl.Solution.Lock(ctx, key, ttl)
		if err == nil {
			return nil
		}

		// 合约不存在，可能已过期，不再重试
		if err != nil && err.Error() == rpctypes.ErrorDesc(rpctypes.ErrGRPCLeaseNotFound) {
			return errors.New(pkg.ErrEtcdV3LockFailure)
		}
	}

	// 有可能网络原因
	if err != nil {
		return err
	}

	return nil
}

func (nl *Nxlock) Release(ctx context.Context, key string) error {
	nl.holdLock.Delete(key)

	err := nl.Solution.Release(ctx, key)
	if err != nil {
		return err
	}

	return nil
}

func (nl *Nxlock) Close() error {
	return nl.Solution.Close()
}
