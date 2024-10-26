// Code generated by mockery v2.46.3. DO NOT EDIT.

package models

import (
	mock "github.com/stretchr/testify/mock"
	models "gophkeeper.com/internal/server/models"
)

// Secret is an autogenerated mock type for the Secret type
type Secret struct {
	mock.Mock
}

type Secret_Expecter struct {
	mock *mock.Mock
}

func (_m *Secret) EXPECT() *Secret_Expecter {
	return &Secret_Expecter{mock: &_m.Mock}
}

// Accept provides a mock function with given fields: visitor
func (_m *Secret) Accept(visitor models.SecretVisitor) error {
	ret := _m.Called(visitor)

	if len(ret) == 0 {
		panic("no return value specified for Accept")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(models.SecretVisitor) error); ok {
		r0 = rf(visitor)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Secret_Accept_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Accept'
type Secret_Accept_Call struct {
	*mock.Call
}

// Accept is a helper method to define mock.On call
//   - visitor models.SecretVisitor
func (_e *Secret_Expecter) Accept(visitor interface{}) *Secret_Accept_Call {
	return &Secret_Accept_Call{Call: _e.mock.On("Accept", visitor)}
}

func (_c *Secret_Accept_Call) Run(run func(visitor models.SecretVisitor)) *Secret_Accept_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(models.SecretVisitor))
	})
	return _c
}

func (_c *Secret_Accept_Call) Return(_a0 error) *Secret_Accept_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Secret_Accept_Call) RunAndReturn(run func(models.SecretVisitor) error) *Secret_Accept_Call {
	_c.Call.Return(run)
	return _c
}

// NewSecret creates a new instance of Secret. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSecret(t interface {
	mock.TestingT
	Cleanup(func())
}) *Secret {
	mock := &Secret{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}