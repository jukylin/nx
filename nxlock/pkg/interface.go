package pkg

import (
	"context"
)

type NxlockSolution interface {
	Lock(ctx context.Context, key string, ttl int64) error

	Release(ctx context.Context, key string) error

	Close() error
}
