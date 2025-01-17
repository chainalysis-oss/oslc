// Code generated by mockery v2.50.1. DO NOT EDIT.

package oslc

import (
	oslc "github.com/chainalysis-oss/oslc"
	mock "github.com/stretchr/testify/mock"
)

// MockLicenseRetriever is an autogenerated mock type for the LicenseRetriever type
type MockLicenseRetriever struct {
	mock.Mock
}

type MockLicenseRetriever_Expecter struct {
	mock *mock.Mock
}

func (_m *MockLicenseRetriever) EXPECT() *MockLicenseRetriever_Expecter {
	return &MockLicenseRetriever_Expecter{mock: &_m.Mock}
}

// Licenses provides a mock function with no fields
func (_m *MockLicenseRetriever) Licenses() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Licenses")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// MockLicenseRetriever_Licenses_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Licenses'
type MockLicenseRetriever_Licenses_Call struct {
	*mock.Call
}

// Licenses is a helper method to define mock.On call
func (_e *MockLicenseRetriever_Expecter) Licenses() *MockLicenseRetriever_Licenses_Call {
	return &MockLicenseRetriever_Licenses_Call{Call: _e.mock.On("Licenses")}
}

func (_c *MockLicenseRetriever_Licenses_Call) Run(run func()) *MockLicenseRetriever_Licenses_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockLicenseRetriever_Licenses_Call) Return(_a0 []string) *MockLicenseRetriever_Licenses_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseRetriever_Licenses_Call) RunAndReturn(run func() []string) *MockLicenseRetriever_Licenses_Call {
	_c.Call.Return(run)
	return _c
}

// Lookup provides a mock function with given fields: id
func (_m *MockLicenseRetriever) Lookup(id string) oslc.License {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Lookup")
	}

	var r0 oslc.License
	if rf, ok := ret.Get(0).(func(string) oslc.License); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(oslc.License)
	}

	return r0
}

// MockLicenseRetriever_Lookup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lookup'
type MockLicenseRetriever_Lookup_Call struct {
	*mock.Call
}

// Lookup is a helper method to define mock.On call
//   - id string
func (_e *MockLicenseRetriever_Expecter) Lookup(id interface{}) *MockLicenseRetriever_Lookup_Call {
	return &MockLicenseRetriever_Lookup_Call{Call: _e.mock.On("Lookup", id)}
}

func (_c *MockLicenseRetriever_Lookup_Call) Run(run func(id string)) *MockLicenseRetriever_Lookup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockLicenseRetriever_Lookup_Call) Return(_a0 oslc.License) *MockLicenseRetriever_Lookup_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseRetriever_Lookup_Call) RunAndReturn(run func(string) oslc.License) *MockLicenseRetriever_Lookup_Call {
	_c.Call.Return(run)
	return _c
}

// ReleaseDate provides a mock function with no fields
func (_m *MockLicenseRetriever) ReleaseDate() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ReleaseDate")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockLicenseRetriever_ReleaseDate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReleaseDate'
type MockLicenseRetriever_ReleaseDate_Call struct {
	*mock.Call
}

// ReleaseDate is a helper method to define mock.On call
func (_e *MockLicenseRetriever_Expecter) ReleaseDate() *MockLicenseRetriever_ReleaseDate_Call {
	return &MockLicenseRetriever_ReleaseDate_Call{Call: _e.mock.On("ReleaseDate")}
}

func (_c *MockLicenseRetriever_ReleaseDate_Call) Run(run func()) *MockLicenseRetriever_ReleaseDate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockLicenseRetriever_ReleaseDate_Call) Return(_a0 string) *MockLicenseRetriever_ReleaseDate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseRetriever_ReleaseDate_Call) RunAndReturn(run func() string) *MockLicenseRetriever_ReleaseDate_Call {
	_c.Call.Return(run)
	return _c
}

// Source provides a mock function with no fields
func (_m *MockLicenseRetriever) Source() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Source")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockLicenseRetriever_Source_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Source'
type MockLicenseRetriever_Source_Call struct {
	*mock.Call
}

// Source is a helper method to define mock.On call
func (_e *MockLicenseRetriever_Expecter) Source() *MockLicenseRetriever_Source_Call {
	return &MockLicenseRetriever_Source_Call{Call: _e.mock.On("Source")}
}

func (_c *MockLicenseRetriever_Source_Call) Run(run func()) *MockLicenseRetriever_Source_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockLicenseRetriever_Source_Call) Return(_a0 string) *MockLicenseRetriever_Source_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseRetriever_Source_Call) RunAndReturn(run func() string) *MockLicenseRetriever_Source_Call {
	_c.Call.Return(run)
	return _c
}

// Version provides a mock function with no fields
func (_m *MockLicenseRetriever) Version() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Version")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockLicenseRetriever_Version_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Version'
type MockLicenseRetriever_Version_Call struct {
	*mock.Call
}

// Version is a helper method to define mock.On call
func (_e *MockLicenseRetriever_Expecter) Version() *MockLicenseRetriever_Version_Call {
	return &MockLicenseRetriever_Version_Call{Call: _e.mock.On("Version")}
}

func (_c *MockLicenseRetriever_Version_Call) Run(run func()) *MockLicenseRetriever_Version_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockLicenseRetriever_Version_Call) Return(_a0 string) *MockLicenseRetriever_Version_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseRetriever_Version_Call) RunAndReturn(run func() string) *MockLicenseRetriever_Version_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockLicenseRetriever creates a new instance of MockLicenseRetriever. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockLicenseRetriever(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLicenseRetriever {
	mock := &MockLicenseRetriever{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
