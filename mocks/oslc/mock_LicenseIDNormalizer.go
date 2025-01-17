// Code generated by mockery v2.50.1. DO NOT EDIT.

package oslc

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockLicenseIDNormalizer is an autogenerated mock type for the LicenseIDNormalizer type
type MockLicenseIDNormalizer struct {
	mock.Mock
}

type MockLicenseIDNormalizer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockLicenseIDNormalizer) EXPECT() *MockLicenseIDNormalizer_Expecter {
	return &MockLicenseIDNormalizer_Expecter{mock: &_m.Mock}
}

// NormalizeID provides a mock function with given fields: ctx, id
func (_m *MockLicenseIDNormalizer) NormalizeID(ctx context.Context, id string) string {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for NormalizeID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockLicenseIDNormalizer_NormalizeID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NormalizeID'
type MockLicenseIDNormalizer_NormalizeID_Call struct {
	*mock.Call
}

// NormalizeID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *MockLicenseIDNormalizer_Expecter) NormalizeID(ctx interface{}, id interface{}) *MockLicenseIDNormalizer_NormalizeID_Call {
	return &MockLicenseIDNormalizer_NormalizeID_Call{Call: _e.mock.On("NormalizeID", ctx, id)}
}

func (_c *MockLicenseIDNormalizer_NormalizeID_Call) Run(run func(ctx context.Context, id string)) *MockLicenseIDNormalizer_NormalizeID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockLicenseIDNormalizer_NormalizeID_Call) Return(_a0 string) *MockLicenseIDNormalizer_NormalizeID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockLicenseIDNormalizer_NormalizeID_Call) RunAndReturn(run func(context.Context, string) string) *MockLicenseIDNormalizer_NormalizeID_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockLicenseIDNormalizer creates a new instance of MockLicenseIDNormalizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockLicenseIDNormalizer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLicenseIDNormalizer {
	mock := &MockLicenseIDNormalizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
