// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	model "kvdb/internal/model"

	mock "github.com/stretchr/testify/mock"
)

// Compute is an autogenerated mock type for the compute type
type Compute struct {
	mock.Mock
}

type Compute_Expecter struct {
	mock *mock.Mock
}

func (_m *Compute) EXPECT() *Compute_Expecter {
	return &Compute_Expecter{mock: &_m.Mock}
}

// Parse provides a mock function with given fields: query
func (_m *Compute) Parse(query string) (model.Query, error) {
	ret := _m.Called(query)

	if len(ret) == 0 {
		panic("no return value specified for Parse")
	}

	var r0 model.Query
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (model.Query, error)); ok {
		return rf(query)
	}
	if rf, ok := ret.Get(0).(func(string) model.Query); ok {
		r0 = rf(query)
	} else {
		r0 = ret.Get(0).(model.Query)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Compute_Parse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Parse'
type Compute_Parse_Call struct {
	*mock.Call
}

// Parse is a helper method to define mock.On call
//   - query string
func (_e *Compute_Expecter) Parse(query interface{}) *Compute_Parse_Call {
	return &Compute_Parse_Call{Call: _e.mock.On("Parse", query)}
}

func (_c *Compute_Parse_Call) Run(run func(query string)) *Compute_Parse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Compute_Parse_Call) Return(_a0 model.Query, _a1 error) *Compute_Parse_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Compute_Parse_Call) RunAndReturn(run func(string) (model.Query, error)) *Compute_Parse_Call {
	_c.Call.Return(run)
	return _c
}

// NewCompute creates a new instance of Compute. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCompute(t interface {
	mock.TestingT
	Cleanup(func())
}) *Compute {
	mock := &Compute{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
