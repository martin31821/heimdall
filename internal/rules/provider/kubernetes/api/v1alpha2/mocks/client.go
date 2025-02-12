// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	v1alpha2 "github.com/dadrus/heimdall/internal/rules/provider/kubernetes/api/v1alpha2"
	mock "github.com/stretchr/testify/mock"
)

// ClientMock is an autogenerated mock type for the Client type
type ClientMock struct {
	mock.Mock
}

type ClientMock_Expecter struct {
	mock *mock.Mock
}

func (_m *ClientMock) EXPECT() *ClientMock_Expecter {
	return &ClientMock_Expecter{mock: &_m.Mock}
}

// RuleSetRepository provides a mock function with given fields: namespace
func (_m *ClientMock) RuleSetRepository(namespace string) v1alpha2.RuleSetRepository {
	ret := _m.Called(namespace)

	var r0 v1alpha2.RuleSetRepository
	if rf, ok := ret.Get(0).(func(string) v1alpha2.RuleSetRepository); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha2.RuleSetRepository)
		}
	}

	return r0
}

// ClientMock_RuleSetRepository_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RuleSetRepository'
type ClientMock_RuleSetRepository_Call struct {
	*mock.Call
}

// RuleSetRepository is a helper method to define mock.On call
//   - namespace string
func (_e *ClientMock_Expecter) RuleSetRepository(namespace interface{}) *ClientMock_RuleSetRepository_Call {
	return &ClientMock_RuleSetRepository_Call{Call: _e.mock.On("RuleSetRepository", namespace)}
}

func (_c *ClientMock_RuleSetRepository_Call) Run(run func(namespace string)) *ClientMock_RuleSetRepository_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ClientMock_RuleSetRepository_Call) Return(_a0 v1alpha2.RuleSetRepository) *ClientMock_RuleSetRepository_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ClientMock_RuleSetRepository_Call) RunAndReturn(run func(string) v1alpha2.RuleSetRepository) *ClientMock_RuleSetRepository_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewClientMock interface {
	mock.TestingT
	Cleanup(func())
}

// NewClientMock creates a new instance of ClientMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClientMock(t mockConstructorTestingTNewClientMock) *ClientMock {
	mock := &ClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
