// Code generated by mockery v2.46.3. DO NOT EDIT.

package models

import (
	mock "github.com/stretchr/testify/mock"
	models "gophkeeper.com/internal/server/models"
)

// NoteOption is an autogenerated mock type for the NoteOption type
type NoteOption struct {
	mock.Mock
}

type NoteOption_Expecter struct {
	mock *mock.Mock
}

func (_m *NoteOption) EXPECT() *NoteOption_Expecter {
	return &NoteOption_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0
func (_m *NoteOption) Execute(_a0 *models.NoteOptions) {
	_m.Called(_a0)
}

// NoteOption_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type NoteOption_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 *models.NoteOptions
func (_e *NoteOption_Expecter) Execute(_a0 interface{}) *NoteOption_Execute_Call {
	return &NoteOption_Execute_Call{Call: _e.mock.On("Execute", _a0)}
}

func (_c *NoteOption_Execute_Call) Run(run func(_a0 *models.NoteOptions)) *NoteOption_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.NoteOptions))
	})
	return _c
}

func (_c *NoteOption_Execute_Call) Return() *NoteOption_Execute_Call {
	_c.Call.Return()
	return _c
}

func (_c *NoteOption_Execute_Call) RunAndReturn(run func(*models.NoteOptions)) *NoteOption_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewNoteOption creates a new instance of NoteOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNoteOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *NoteOption {
	mock := &NoteOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}