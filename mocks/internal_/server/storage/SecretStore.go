// Code generated by mockery v2.46.3. DO NOT EDIT.

package storage

import mock "github.com/stretchr/testify/mock"

// SecretStore is an autogenerated mock type for the SecretStore type
type SecretStore struct {
	mock.Mock
}

type SecretStore_Expecter struct {
	mock *mock.Mock
}

func (_m *SecretStore) EXPECT() *SecretStore_Expecter {
	return &SecretStore_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: path
func (_m *SecretStore) Delete(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretStore_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type SecretStore_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - path string
func (_e *SecretStore_Expecter) Delete(path interface{}) *SecretStore_Delete_Call {
	return &SecretStore_Delete_Call{Call: _e.mock.On("Delete", path)}
}

func (_c *SecretStore_Delete_Call) Run(run func(path string)) *SecretStore_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SecretStore_Delete_Call) Return(_a0 error) *SecretStore_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretStore_Delete_Call) RunAndReturn(run func(string) error) *SecretStore_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Retrieve provides a mock function with given fields: path
func (_m *SecretStore) Retrieve(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for Retrieve")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretStore_Retrieve_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Retrieve'
type SecretStore_Retrieve_Call struct {
	*mock.Call
}

// Retrieve is a helper method to define mock.On call
//   - path string
func (_e *SecretStore_Expecter) Retrieve(path interface{}) *SecretStore_Retrieve_Call {
	return &SecretStore_Retrieve_Call{Call: _e.mock.On("Retrieve", path)}
}

func (_c *SecretStore_Retrieve_Call) Run(run func(path string)) *SecretStore_Retrieve_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SecretStore_Retrieve_Call) Return(_a0 error) *SecretStore_Retrieve_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretStore_Retrieve_Call) RunAndReturn(run func(string) error) *SecretStore_Retrieve_Call {
	_c.Call.Return(run)
	return _c
}

// Store provides a mock function with given fields: path
func (_m *SecretStore) Store(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SecretStore_Store_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Store'
type SecretStore_Store_Call struct {
	*mock.Call
}

// Store is a helper method to define mock.On call
//   - path string
func (_e *SecretStore_Expecter) Store(path interface{}) *SecretStore_Store_Call {
	return &SecretStore_Store_Call{Call: _e.mock.On("Store", path)}
}

func (_c *SecretStore_Store_Call) Run(run func(path string)) *SecretStore_Store_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SecretStore_Store_Call) Return(_a0 error) *SecretStore_Store_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SecretStore_Store_Call) RunAndReturn(run func(string) error) *SecretStore_Store_Call {
	_c.Call.Return(run)
	return _c
}

// NewSecretStore creates a new instance of SecretStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSecretStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *SecretStore {
	mock := &SecretStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
