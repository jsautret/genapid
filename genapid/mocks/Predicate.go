// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	ctx "github.com/jsautret/genapid/ctx"

	mock "github.com/stretchr/testify/mock"

	zerolog "github.com/rs/zerolog"
)

// Predicate is an autogenerated mock type for the Predicate type
type Predicate struct {
	mock.Mock
}

// Call provides a mock function with given fields: _a0, _a1
func (_m *Predicate) Call(_a0 zerolog.Logger, _a1 *ctx.Ctx) bool {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(zerolog.Logger, *ctx.Ctx) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *Predicate) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Params provides a mock function with given fields:
func (_m *Predicate) Params() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Result provides a mock function with given fields:
func (_m *Predicate) Result() ctx.Result {
	ret := _m.Called()

	var r0 ctx.Result
	if rf, ok := ret.Get(0).(func() ctx.Result); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ctx.Result)
		}
	}

	return r0
}
