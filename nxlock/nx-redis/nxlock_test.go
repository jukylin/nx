package nx_redis

import (
	"testing"
	"os"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/esim/log"
	"context"
	"github.com/jukylin/esim/config"
	"github.com/stretchr/testify/assert"
)

var logger log.Logger
var conf config.Config

func TestMain(m *testing.M) {
	logger = log.NewLogger(
		log.WithDebug(true),
	)

	conf = config.NewMemConfig()
	conf.Set("debug", true)
	code := m.Run()


	os.Exit(code)
}

func TestRedisClient_Lock(t *testing.T) {
	clientOptions := redis.ClientOptions{}
	client := redis.NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithConf(conf),
	)

	rclient := NewClient(
		WithLogger(logger),
		WithClient(client),
	)
	ctx := context.Background()
	key := "TestRedisClient_Lock"
	err := rclient.Lock(ctx, key, "1", 10)
	assert.Nil(t, err)
	err = rclient.Release(ctx, key)
	assert.Nil(t, err)
}