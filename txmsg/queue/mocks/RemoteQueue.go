// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	queue "github.com/jukylin/nx/txmsg/queue"

	mock "github.com/stretchr/testify/mock"
)

// RemoteQueue is an autogenerated mock type for the RemoteQueue type
type RemoteQueue struct {
	mock.Mock
}

// Send provides a mock function with given fields: ctx, msg
func (_m *RemoteQueue) Send(ctx context.Context, msg queue.Message) error {
	ret := _m.Called(ctx, msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, queue.Message) error); ok {
		r0 = rf(ctx, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
