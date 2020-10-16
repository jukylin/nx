package nx_redis

import (
	"context"
	"os"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/redis"
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
	err := rclient.Lock(ctx, key, 10)
	assert.Nil(t, err)
	err = rclient.Release(ctx, key)
	assert.Nil(t, err)

	conn := client.GetCtxRedisConn()
	res, nil := redis.Bool(conn.Do(context.Background(), "exists", key))
	assert.Nil(t, nil)
	assert.False(t, res)
}



func TestRedisClient_keepAliveKey(t *testing.T) {
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
	key := "TestRedisClient_keepAliveKey"
	err := rclient.Lock(ctx, key, 10)
	assert.Nil(t, err)
	_, ok := rclient.(*Client).keepAliveKey[key];
	assert.True(t, ok)

	err = rclient.Release(ctx, key)
	assert.Nil(t, err)
	_, ok = rclient.(*Client).keepAliveKey[key];
	assert.False(t, ok)

	conn := client.GetCtxRedisConn()
	res, nil := redis.Bool(conn.Do(context.Background(), "exists", key))
	assert.Nil(t, nil)
	assert.False(t, res)
}