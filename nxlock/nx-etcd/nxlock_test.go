package nx_etcd

import (
	"context"
	"testing"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/nxlock/nx-etcd/mocks"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/v3/clientv3"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"127.0.0.1:2379"}
)

func TestEtcdV3_LeaseNotFound(t *testing.T) {
	etcdv3 := NewEtcdV3(
		WithConfig(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: dialTimeout,
		}),
		WithLogger(log.NewLogger(
			log.WithDebug(true),
		)),
	)

	ctx := context.Background()

	lease := &mocks.Lease{}
	lease.On("Grant", ctx, int64(10)).Return(&clientv3.LeaseGrantResponse{
		ID:  clientv3.LeaseID(100),
		TTL: 10,
	}, nil)

	lkar := make(<-chan *clientv3.LeaseKeepAliveResponse, 1)
	lease.On("KeepAlive", ctx, clientv3.LeaseID(100)).Return(lkar, nil)
	lease.On("Revoke", ctx, clientv3.LeaseID(100)).Return(&clientv3.LeaseRevokeResponse{}, nil)
	lease.On("Close").Return(nil)
	etcdv3.(*EtcdV3).Lease = lease

	key := "LeaseNotFound"
	err := etcdv3.Lock(ctx, key, "1", 10)
	assert.Error(t, err)
}

func TestEtcdV3_LockAndRelease(t *testing.T) {
	etcdv3 := NewEtcdV3(
		WithConfig(clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: dialTimeout,
		}),
		WithLogger(log.NewLogger(
			log.WithDebug(true),
		)),
	)

	ctx := context.Background()
	key := "LockAndRelease"
	for i := 0; i < 3; i++ {
		err := etcdv3.Lock(ctx, key, "1", 10)
		assert.Nil(t, err)
		err = etcdv3.Release(ctx, key)
		assert.Nil(t, err)
	}
}
