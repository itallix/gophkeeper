// Code generated by mockery v2.46.3. DO NOT EDIT.

package models

import (
	mock "github.com/stretchr/testify/mock"
	models "gophkeeper.com/internal/server/models"
)

// BinaryOption is an autogenerated mock type for the BinaryOption type
type BinaryOption struct {
	mock.Mock
}

type BinaryOption_Expecter struct {
	mock *mock.Mock
}

func (_m *BinaryOption) EXPECT() *BinaryOption_Expecter {
	return &BinaryOption_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *BinaryOption) Execute(_a0 *models.BinaryOptions) {
	_m.Called(_a0)
}

// BinaryOption_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type BinaryOption_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 *models.BinaryOptions
func (_e *BinaryOption_Expecter) Execute(_a0 interface{}) *BinaryOption_Execute_Call {
	return &BinaryOption_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *BinaryOption_Execute_Call) Run(run func(_a0 *models.BinaryOptions)) *BinaryOption_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.BinaryOptions))
	})
	return _c
}

func (_c *BinaryOption_Execute_Call) Return() *BinaryOption_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *BinaryOption_Execute_Call) RunAndReturn(run func(*models.BinaryOptions)) *BinaryOption_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewBinaryOption creates a new instance of BinaryOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBinaryOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *BinaryOption {
	mock := &BinaryOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
