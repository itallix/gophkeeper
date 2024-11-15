// Code generated by mockery v2.46.3. DO NOT EDIT.

package models

import (
	mock "github.com/stretchr/testify/mock"

	models "github.com/itallix/gophkeeper/internal/server/models"
)

// SecretVisitor is an autogenerated mock type for the SecretVisitor type
type SecretVisitor struct {
	mock.Mock
}

type SecretVisitor_Expecter struct {
	mock *mock.Mock
}

func (_m *SecretVisitor) EXPECT() *SecretVisitor_Expecter {
	return &SecretVisitor_Expecter{mock: &_m.Mock}
}

// GetResult provides a mock function with given fields:
func (_m *SecretVisitor) GetResult() any {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetResult")
	}

	var r0 any
	if rf, ok := ret.Get(0).(func() any); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(any)
		}
	}

	return r0
}

// SecretVisitor_GetResult_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetResult'
type SecretVisitor_GetResult_Call struct {
	*mock.Call
}

// GetResult is a helper method to define mock.On call
func (_e *SecretVisitor_Expecter) GetResult() *SecretVisitor_GetResult_Call {
	return &SecretVisitor_GetResult_Call{Call: _e.mock.On("GetResult")}
}

func (_c *SecretVisitor_GetResult_Call) Run(run func()) *SecretVisitor_GetResult_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SecretVisitor_GetResult_Call) Return(_a0 any) *SecretVisitor_GetResult_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretVisitor_GetResult_Call) RunAndReturn(run func() any) *SecretVisitor_GetResult_Call {
	_c.Call.Return(run)
	return _c
}

// VisitBinary provides a mock function with given fields: binary
func (_m *SecretVisitor) VisitBinary(binary *models.Binary) error {
	ret := _m.Called(binary)

	if len(ret) == 0 {
		panic("no return value specified for VisitBinary")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Binary) error); ok {
		r0 = rf(binary)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretVisitor_VisitBinary_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VisitBinary'
type SecretVisitor_VisitBinary_Call struct {
	*mock.Call
}

// VisitBinary is a helper method to define mock.On call
//   - binary *models.Binary
func (_e *SecretVisitor_Expecter) VisitBinary(binary interface{}) *SecretVisitor_VisitBinary_Call {
	return &SecretVisitor_VisitBinary_Call{Call: _e.mock.On("VisitBinary", binary)}
}

func (_c *SecretVisitor_VisitBinary_Call) Run(run func(binary *models.Binary)) *SecretVisitor_VisitBinary_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.Binary))
	})
	return _c
}

func (_c *SecretVisitor_VisitBinary_Call) Return(_a0 error) *SecretVisitor_VisitBinary_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretVisitor_VisitBinary_Call) RunAndReturn(run func(*models.Binary) error) *SecretVisitor_VisitBinary_Call {
	_c.Call.Return(run)
	return _c
}

// VisitCard provides a mock function with given fields: card
func (_m *SecretVisitor) VisitCard(card *models.Card) error {
	ret := _m.Called(card)

	if len(ret) == 0 {
		panic("no return value specified for VisitCard")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Card) error); ok {
		r0 = rf(card)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretVisitor_VisitCard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VisitCard'
type SecretVisitor_VisitCard_Call struct {
	*mock.Call
}

// VisitCard is a helper method to define mock.On call
//   - card *models.Card
func (_e *SecretVisitor_Expecter) VisitCard(card interface{}) *SecretVisitor_VisitCard_Call {
	return &SecretVisitor_VisitCard_Call{Call: _e.mock.On("VisitCard", card)}
}

func (_c *SecretVisitor_VisitCard_Call) Run(run func(card *models.Card)) *SecretVisitor_VisitCard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.Card))
	})
	return _c
}

func (_c *SecretVisitor_VisitCard_Call) Return(_a0 error) *SecretVisitor_VisitCard_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretVisitor_VisitCard_Call) RunAndReturn(run func(*models.Card) error) *SecretVisitor_VisitCard_Call {
	_c.Call.Return(run)
	return _c
}

// VisitLogin provides a mock function with given fields: login
func (_m *SecretVisitor) VisitLogin(login *models.Login) error {
	ret := _m.Called(login)

	if len(ret) == 0 {
		panic("no return value specified for VisitLogin")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Login) error); ok {
		r0 = rf(login)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretVisitor_VisitLogin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VisitLogin'
type SecretVisitor_VisitLogin_Call struct {
	*mock.Call
}

// VisitLogin is a helper method to define mock.On call
//   - login *models.Login
func (_e *SecretVisitor_Expecter) VisitLogin(login interface{}) *SecretVisitor_VisitLogin_Call {
	return &SecretVisitor_VisitLogin_Call{Call: _e.mock.On("VisitLogin", login)}
}

func (_c *SecretVisitor_VisitLogin_Call) Run(run func(login *models.Login)) *SecretVisitor_VisitLogin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.Login))
	})
	return _c
}

func (_c *SecretVisitor_VisitLogin_Call) Return(_a0 error) *SecretVisitor_VisitLogin_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretVisitor_VisitLogin_Call) RunAndReturn(run func(*models.Login) error) *SecretVisitor_VisitLogin_Call {
	_c.Call.Return(run)
	return _c
}

// VisitNote provides a mock function with given fields: note
func (_m *SecretVisitor) VisitNote(note *models.Note) error {
	ret := _m.Called(note)

	if len(ret) == 0 {
		panic("no return value specified for VisitNote")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Note) error); ok {
		r0 = rf(note)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretVisitor_VisitNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VisitNote'
type SecretVisitor_VisitNote_Call struct {
	*mock.Call
}

// VisitNote is a helper method to define mock.On call
//   - note *models.Note
func (_e *SecretVisitor_Expecter) VisitNote(note interface{}) *SecretVisitor_VisitNote_Call {
	return &SecretVisitor_VisitNote_Call{Call: _e.mock.On("VisitNote", note)}
}

func (_c *SecretVisitor_VisitNote_Call) Run(run func(note *models.Note)) *SecretVisitor_VisitNote_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*models.Note))
	})
	return _c
}

func (_c *SecretVisitor_VisitNote_Call) Return(_a0 error) *SecretVisitor_VisitNote_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretVisitor_VisitNote_Call) RunAndReturn(run func(*models.Note) error) *SecretVisitor_VisitNote_Call {
	_c.Call.Return(run)
	return _c
}

// NewSecretVisitor creates a new instance of SecretVisitor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSecretVisitor(t interface {
	mock.TestingT
	Cleanup(func())
}) *SecretVisitor {
	mock := &SecretVisitor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
