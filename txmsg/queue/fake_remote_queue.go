package queue

import (
	"context"
)

type FakeRemoteQueue struct {
}

// Send provides a mock function with given fields: ctx, msg
func (f *FakeRemoteQueue) Send(ctx context.Context, msg Message) error {
	return nil
}
