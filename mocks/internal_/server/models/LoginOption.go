// Code generated by mockery v2.46.3. DO NOT EDIT.

package models

import (
	mock "github.com/stretchr/testify/mock"

	models "github.com/itallix/gophkeeper/internal/server/models"
)

// LoginOption is an autogenerated mock type for the LoginOption type
type LoginOption struct {
	mock.Mock
}

type LoginOption_Expecter struct {
	mock *mock.Mock
}

func (_m *LoginOption) EXPECT() *LoginOption_Expecter {
	return &LoginOption_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *LoginOption) Execute(_a0 *models.LoginOptions) {
	_m.Called(_a0)
}

// LoginOption_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type LoginOption_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 *models.LoginOptions
func (_e *LoginOption_Expecter) Execute(_a0 interface{}) *LoginOption_Execute_Call {
	return &LoginOption_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *LoginOption_Execute_Call) Run(run func(_a0 *models.LoginOptions)) *LoginOption_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.LoginOptions))
	})
	return _c
}

func (_c *LoginOption_Execute_Call) Return() *LoginOption_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *LoginOption_Execute_Call) RunAndReturn(run func(*models.LoginOptions)) *LoginOption_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewLoginOption creates a new instance of LoginOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLoginOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *LoginOption {
	mock := &LoginOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
