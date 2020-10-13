// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/jukylin/nx/saga/domain/entity"
	mock "github.com/stretchr/testify/mock"
)

// Compensate is an autogenerated mock type for the Compensate type
type Compensate struct {
	mock.Mock
}

// BuildCompensate provides a mock function with given fields: ctx, txgroup
func (_m *Compensate) BuildCompensate(ctx context.Context, txgroup entity.Txgroup) error {
	ret := _m.Called(ctx, txgroup)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entity.Txgroup) error); ok {
		r0 = rf(ctx, txgroup)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CompensateHook provides a mock function with given fields: ctx
func (_m *Compensate) CompensateHook(ctx context.Context) {
	_m.Called(ctx)
}

// CompensateRecord provides a mock function with given fields: ctx
func (_m *Compensate) CompensateRecord(ctx context.Context) {
	_m.Called(ctx)
}

// ExeCompensate provides a mock function with given fields: ctx, txgroup
func (_m *Compensate) ExeCompensate(ctx context.Context, txgroup entity.Txgroup) {
	_m.Called(ctx, txgroup)
}