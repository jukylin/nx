package nx_redis

import (
	"context"
	"errors"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/nx/nxlock/pkg"
)

type Client struct {
	*redis.Client

	logger log.Logger
}

type ClientOption func(*Client)

// 分布式锁 redis > 2.6 解决方案
func NewClient(options ...ClientOption) pkg.NxlockSolution {
	rc := &Client{}

	for _, option := range options {
		option(rc)
	}

	return rc
}

func WithLogger(logger log.Logger) ClientOption {
	return func(e *Client) {
		e.logger = logger
	}
}

func WithClient(client *redis.Client) ClientOption {
	return func(e *Client) {
		e.Client = client
	}
}

func (rc *Client) Lock(ctx context.Context, key string, ttl int64) error {
	err := rc.set(ctx, key, "1", ttl)
	if err != nil {
		rc.logger.Debugc(ctx, err.Error())
		return err
	}

	go rc.keepAlive(ctx, key, ttl)

	return nil
}

func (rc *Client) set(ctx context.Context, key, val string, ttl int64) error {
	conn := rc.GetCtxRedisConn()
	defer conn.Close()

	ok, err := redis.String(conn.Do(ctx, "set", key, val, "nx", "ex", ttl))
	if err != nil {
		return err
	}

	if ok != "OK" {
		return errors.New(pkg.ErrRedisLockFailure)
	}

	return err
}

func (rc *Client) Release(ctx context.Context, key string) error {
	return rc.expire(ctx, key, -1)
}

// 续租
func (rc *Client) keepAlive(ctx context.Context, key string, ttl int64) {
	ticker := time.NewTicker(time.Duration(ttl/3) * time.Second)

	for {
		select {
		case <-ticker.C:
			err := rc.expire(ctx, key, ttl)
			if err != nil {
				rc.logger.Errorc(ctx, "续租失败 %s", err.Error())
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (rc *Client) expire(ctx context.Context, key string, ttl int64) error {
	conn := rc.GetCtxRedisConn()
	defer conn.Close()

	_, err := redis.Bool(conn.Do(ctx, "expire", key, ttl))
	if err != nil {
		return err
	}

	return nil
}

func (rc *Client) Close() error {
	return rc.Client.Close()
}
