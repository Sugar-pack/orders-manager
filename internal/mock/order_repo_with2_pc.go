// Code generated by mockery v2.12.1. DO NOT EDIT.

package mock

import (
	context "context"

	repository "github.com/Sugar-pack/orders-manager/internal/repository"
	mock "github.com/stretchr/testify/mock"

	testing "testing"

	uuid "github.com/google/uuid"
)

// OrderRepoWith2PC is an autogenerated mock type for the OrderRepoWith2PC type
type OrderRepoWith2PC struct {
	mock.Mock
}

// CommitInsertTransaction provides a mock function with given fields: ctx, txID
func (_m *OrderRepoWith2PC) CommitInsertTransaction(ctx context.Context, txID uuid.UUID) error {
	ret := _m.Called(ctx, txID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, txID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetOrder provides a mock function with given fields: ctx, id
func (_m *OrderRepoWith2PC) GetOrder(ctx context.Context, id uuid.UUID) (*repository.Order, error) {
	ret := _m.Called(ctx, id)

	var r0 *repository.Order
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *repository.Order); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareInsertOrder provides a mock function with given fields: ctx, order, txId
func (_m *OrderRepoWith2PC) PrepareInsertOrder(ctx context.Context, order *repository.Order, txId uuid.UUID) error {
	ret := _m.Called(ctx, order, txId)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *repository.Order, uuid.UUID) error); ok {
		r0 = rf(ctx, order, txId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RollbackInsertTransaction provides a mock function with given fields: ctx, txID
func (_m *OrderRepoWith2PC) RollbackInsertTransaction(ctx context.Context, txID uuid.UUID) error {
	ret := _m.Called(ctx, txID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, txID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewOrderRepoWith2PC creates a new instance of OrderRepoWith2PC. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewOrderRepoWith2PC(t testing.TB) *OrderRepoWith2PC {
	mock := &OrderRepoWith2PC{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
