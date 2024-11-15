// Code generated by mockery v2.46.3. DO NOT EDIT.

package service

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// EncryptionService is an autogenerated mock type for the EncryptionService type
type EncryptionService struct {
	mock.Mock
}

type EncryptionService_Expecter struct {
	mock *mock.Mock
}

func (_m *EncryptionService) EXPECT() *EncryptionService_Expecter {
	return &EncryptionService_Expecter{mock: &_m.Mock}
}

// Decrypt provides a mock function with given fields: src, dst, encryptedDataKey
func (_m *EncryptionService) Decrypt(src []byte, dst io.Writer, encryptedDataKey []byte) error {
	ret := _m.Called(src, dst, encryptedDataKey)

	if len(ret) == 0 {
		panic("no return value specified for Decrypt")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, io.Writer, []byte) error); ok {
		r0 = rf(src, dst, encryptedDataKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EncryptionService_Decrypt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Decrypt'
type EncryptionService_Decrypt_Call struct {
	*mock.Call
}

// Decrypt is a helper method to define mock.On call
//   - src []byte
//   - dst io.Writer
//   - encryptedDataKey []byte
func (_e *EncryptionService_Expecter) Decrypt(src interface{}, dst interface{}, encryptedDataKey interface{}) *EncryptionService_Decrypt_Call {
	return &EncryptionService_Decrypt_Call{Call: _e.mock.On("Decrypt", src, dst, encryptedDataKey)}
}

func (_c *EncryptionService_Decrypt_Call) Run(run func(src []byte, dst io.Writer, encryptedDataKey []byte)) *EncryptionService_Decrypt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte), args[1].(io.Writer), args[2].([]byte))
	})
	return _c
}

func (_c *EncryptionService_Decrypt_Call) Return(_a0 error) *EncryptionService_Decrypt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EncryptionService_Decrypt_Call) RunAndReturn(run func([]byte, io.Writer, []byte) error) *EncryptionService_Decrypt_Call {
	_c.Call.Return(run)
	return _c
}

// Encrypt provides a mock function with given fields: src, dst
func (_m *EncryptionService) Encrypt(src []byte, dst io.Writer) ([]byte, error) {
	ret := _m.Called(src, dst)

	if len(ret) == 0 {
		panic("no return value specified for Encrypt")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte, io.Writer) ([]byte, error)); ok {
		return rf(src, dst)
	}
	if rf, ok := ret.Get(0).(func([]byte, io.Writer) []byte); ok {
		r0 = rf(src, dst)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func([]byte, io.Writer) error); ok {
		r1 = rf(src, dst)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EncryptionService_Encrypt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Encrypt'
type EncryptionService_Encrypt_Call struct {
	*mock.Call
}

// Encrypt is a helper method to define mock.On call
//   - src []byte
//   - dst io.Writer
func (_e *EncryptionService_Expecter) Encrypt(src interface{}, dst interface{}) *EncryptionService_Encrypt_Call {
	return &EncryptionService_Encrypt_Call{Call: _e.mock.On("Encrypt", src, dst)}
}

func (_c *EncryptionService_Encrypt_Call) Run(run func(src []byte, dst io.Writer)) *EncryptionService_Encrypt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte), args[1].(io.Writer))
	})
	return _c
}

func (_c *EncryptionService_Encrypt_Call) Return(_a0 []byte, _a1 error) *EncryptionService_Encrypt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EncryptionService_Encrypt_Call) RunAndReturn(run func([]byte, io.Writer) ([]byte, error)) *EncryptionService_Encrypt_Call {
	_c.Call.Return(run)
	return _c
}

// EncryptWithKey provides a mock function with given fields: src, dst, encryptedDataKey
func (_m *EncryptionService) EncryptWithKey(src []byte, dst io.Writer, encryptedDataKey []byte) error {
	ret := _m.Called(src, dst, encryptedDataKey)

	if len(ret) == 0 {
		panic("no return value specified for EncryptWithKey")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, io.Writer, []byte) error); ok {
		r0 = rf(src, dst, encryptedDataKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EncryptionService_EncryptWithKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EncryptWithKey'
type EncryptionService_EncryptWithKey_Call struct {
	*mock.Call
}

// EncryptWithKey is a helper method to define mock.On call
//   - src []byte
//   - dst io.Writer
//   - encryptedDataKey []byte
func (_e *EncryptionService_Expecter) EncryptWithKey(src interface{}, dst interface{}, encryptedDataKey interface{}) *EncryptionService_EncryptWithKey_Call {
	return &EncryptionService_EncryptWithKey_Call{Call: _e.mock.On("EncryptWithKey", src, dst, encryptedDataKey)}
}

func (_c *EncryptionService_EncryptWithKey_Call) Run(run func(src []byte, dst io.Writer, encryptedDataKey []byte)) *EncryptionService_EncryptWithKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte), args[1].(io.Writer), args[2].([]byte))
	})
	return _c
}

func (_c *EncryptionService_EncryptWithKey_Call) Return(_a0 error) *EncryptionService_EncryptWithKey_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EncryptionService_EncryptWithKey_Call) RunAndReturn(run func([]byte, io.Writer, []byte) error) *EncryptionService_EncryptWithKey_Call {
	_c.Call.Return(run)
	return _c
}

// NewEncryptionService creates a new instance of EncryptionService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEncryptionService(t interface {
	mock.TestingT
	Cleanup(func())
}) *EncryptionService {
	mock := &EncryptionService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
