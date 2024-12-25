// Code generated by mockery v2.46.3. DO NOT EDIT.

package oslc

import (
	oslc "github.com/chainalysis-oss/oslc"
	mock "github.com/stretchr/testify/mock"
)

// MockDistributorClient is an autogenerated mock type for the DistributorClient type
type MockDistributorClient struct {
	mock.Mock
}

type MockDistributorClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDistributorClient) EXPECT() *MockDistributorClient_Expecter {
	return &MockDistributorClient_Expecter{mock: &_m.Mock}
}

// GetPackage provides a mock function with given fields: name
func (_m *MockDistributorClient) GetPackage(name string) (oslc.Entry, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for GetPackage")
	}

	var r0 oslc.Entry
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (oslc.Entry, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) oslc.Entry); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(oslc.Entry)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDistributorClient_GetPackage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPackage'
type MockDistributorClient_GetPackage_Call struct {
	*mock.Call
}

// GetPackage is a helper method to define mock.On call
//   - name string
func (_e *MockDistributorClient_Expecter) GetPackage(name interface{}) *MockDistributorClient_GetPackage_Call {
	return &MockDistributorClient_GetPackage_Call{Call: _e.mock.On("GetPackage", name)}
}

func (_c *MockDistributorClient_GetPackage_Call) Run(run func(name string)) *MockDistributorClient_GetPackage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockDistributorClient_GetPackage_Call) Return(_a0 oslc.Entry, _a1 error) *MockDistributorClient_GetPackage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDistributorClient_GetPackage_Call) RunAndReturn(run func(string) (oslc.Entry, error)) *MockDistributorClient_GetPackage_Call {
	_c.Call.Return(run)
	return _c
}

// GetPackageVersion provides a mock function with given fields: name, version
func (_m *MockDistributorClient) GetPackageVersion(name string, version string) (oslc.Entry, error) {
	ret := _m.Called(name, version)

	if len(ret) == 0 {
		panic("no return value specified for GetPackageVersion")
	}

	var r0 oslc.Entry
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (oslc.Entry, error)); ok {
		return rf(name, version)
	}
	if rf, ok := ret.Get(0).(func(string, string) oslc.Entry); ok {
		r0 = rf(name, version)
	} else {
		r0 = ret.Get(0).(oslc.Entry)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(name, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDistributorClient_GetPackageVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPackageVersion'
type MockDistributorClient_GetPackageVersion_Call struct {
	*mock.Call
}

// GetPackageVersion is a helper method to define mock.On call
//   - name string
//   - version string
func (_e *MockDistributorClient_Expecter) GetPackageVersion(name interface{}, version interface{}) *MockDistributorClient_GetPackageVersion_Call {
	return &MockDistributorClient_GetPackageVersion_Call{Call: _e.mock.On("GetPackageVersion", name, version)}
}

func (_c *MockDistributorClient_GetPackageVersion_Call) Run(run func(name string, version string)) *MockDistributorClient_GetPackageVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockDistributorClient_GetPackageVersion_Call) Return(_a0 oslc.Entry, _a1 error) *MockDistributorClient_GetPackageVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDistributorClient_GetPackageVersion_Call) RunAndReturn(run func(string, string) (oslc.Entry, error)) *MockDistributorClient_GetPackageVersion_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDistributorClient creates a new instance of MockDistributorClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDistributorClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDistributorClient {
	mock := &MockDistributorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}