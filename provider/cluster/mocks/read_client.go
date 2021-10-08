// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	context "context"

	cluster "github.com/virtengine/virtengine/provider/cluster/types"

	mock "github.com/stretchr/testify/mock"

	types "github.com/virtengine/virtengine/x/market/types"
)

// ReadClient is an autogenerated mock type for the ReadClient type
type ReadClient struct {
	mock.Mock
}

// LeaseEvents provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *ReadClient) LeaseEvents(_a0 context.Context, _a1 types.LeaseID, _a2 string, _a3 bool) (cluster.EventsWatcher, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 cluster.EventsWatcher
	if rf, ok := ret.Get(0).(func(context.Context, types.LeaseID, string, bool) cluster.EventsWatcher); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cluster.EventsWatcher)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.LeaseID, string, bool) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LeaseLogs provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *ReadClient) LeaseLogs(_a0 context.Context, _a1 types.LeaseID, _a2 string, _a3 bool, _a4 *int64) ([]*cluster.ServiceLog, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	var r0 []*cluster.ServiceLog
	if rf, ok := ret.Get(0).(func(context.Context, types.LeaseID, string, bool, *int64) []*cluster.ServiceLog); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*cluster.ServiceLog)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.LeaseID, string, bool, *int64) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LeaseStatus provides a mock function with given fields: _a0, _a1
func (_m *ReadClient) LeaseStatus(_a0 context.Context, _a1 types.LeaseID) (*cluster.LeaseStatus, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *cluster.LeaseStatus
	if rf, ok := ret.Get(0).(func(context.Context, types.LeaseID) *cluster.LeaseStatus); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cluster.LeaseStatus)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.LeaseID) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceStatus provides a mock function with given fields: _a0, _a1, _a2
func (_m *ReadClient) ServiceStatus(_a0 context.Context, _a1 types.LeaseID, _a2 string) (*cluster.ServiceStatus, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *cluster.ServiceStatus
	if rf, ok := ret.Get(0).(func(context.Context, types.LeaseID, string) *cluster.ServiceStatus); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cluster.ServiceStatus)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.LeaseID, string) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}