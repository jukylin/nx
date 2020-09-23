package nxlock

import (
	"context"
	"sync"
	"testing"
	"os"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
	"github.com/stretchr/testify/assert"
	"study-go/nxlock/nx-redis"
	"github.com/jukylin/esim/redis"
	"study-go/nxlock/pkg"
)

var logger log.Logger
var conf config.Config
var redisClient *redis.Client
var nxRedisClient pkg.NxlockSolution

func TestMain(m *testing.M) {
	logger = log.NewLogger(
		log.WithDebug(true),
	)

	conf = config.NewMemConfig()
	conf.Set("debug", true)

	clientOptions := redis.ClientOptions{}
	redisClient = redis.NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithConf(conf),
	)

	nxRedisClient = nx_redis.NewClient(
		nx_redis.WithLogger(logger),
		nx_redis.WithClient(redisClient),
	)

	code := m.Run()


	os.Exit(code)
}

func TestNxlock_LocalGoroutineLock(t *testing.T) {
	nxlock := NewNxlock(
		WithSolution(nxRedisClient),
		WithLogger(logger),
	)

	ctx := context.Background()
	wg := sync.WaitGroup{}
	key := "LocalGoroutineLock"
	wg.Add(2)
	go func() {
		err := nxlock.Lock(ctx, key, "1", 10)
		assert.Nil(t, err)
		wg.Done()
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := nxlock.Lock(ctx, key, "1", 10)
		assert.Error(t, err)
		wg.Done()
	}()
	wg.Wait()
	nxlock.Release(ctx, key)
}


func TestNxlock_RedisSolution_SimulationMulProcessLock(t *testing.T) {
	ctx := context.Background()
	var e1 pkg.NxlockSolution
	var e2 pkg.NxlockSolution
	var err error

	wg := sync.WaitGroup{}
	wg.Add(2)
	key := "RedisSolution_SimulationMulProcessLock"
	go func() {
		e1 = NewNxlock(
			WithSolution(nxRedisClient),
			WithLogger(logger),
		)
		err = e1.Lock(ctx, key, "1", 10)
		assert.Nil(t, err)
		wg.Done()
	}()

	go func() {
		e2 = NewNxlock(
			WithSolution(nxRedisClient),
			WithLogger(logger),
		)
		time.Sleep(200 * time.Millisecond)
		err := e2.Lock(ctx, key, "1", 10)
		assert.Equal(t, pkg.ErrRedisLockFailure, err.Error())
		wg.Done()
	}()
	wg.Wait()

	err = e1.Release(ctx, key)
	assert.Nil(t, err)
}