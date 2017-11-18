// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces (interfaces: Trigger)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockTrigger is a mock of Trigger interface
type MockTrigger struct {
	ctrl     *gomock.Controller
	recorder *MockTriggerMockRecorder
}

// MockTriggerMockRecorder is the mock recorder for MockTrigger
type MockTriggerMockRecorder struct {
	mock *MockTrigger
}

// NewMockTrigger creates a new mock instance
func NewMockTrigger(ctrl *gomock.Controller) *MockTrigger {
	mock := &MockTrigger{ctrl: ctrl}
	mock.recorder = &MockTriggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTrigger) EXPECT() *MockTriggerMockRecorder {
	return m.recorder
}

// Activate mocks base method
func (m *MockTrigger) Activate() {
	m.ctrl.Call(m, "Activate")
}

// Activate indicates an expected call of Activate
func (mr *MockTriggerMockRecorder) Activate() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Activate", reflect.TypeOf((*MockTrigger)(nil).Activate))
}

// Completed mocks base method
func (m *MockTrigger) Completed() bool {
	ret := m.ctrl.Call(m, "Completed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Completed indicates an expected call of Completed
func (mr *MockTriggerMockRecorder) Completed() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Completed", reflect.TypeOf((*MockTrigger)(nil).Completed))
}

// Delete mocks base method
func (m *MockTrigger) Delete() error {
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockTriggerMockRecorder) Delete() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockTrigger)(nil).Delete))
}

// Replicas mocks base method
func (m *MockTrigger) Replicas() int {
	ret := m.ctrl.Call(m, "Replicas")
	ret0, _ := ret[0].(int)
	return ret0
}

// Replicas indicates an expected call of Replicas
func (mr *MockTriggerMockRecorder) Replicas() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Replicas", reflect.TypeOf((*MockTrigger)(nil).Replicas))
}

// WaitOn mocks base method
func (m *MockTrigger) WaitOn() (bool, error) {
	ret := m.ctrl.Call(m, "WaitOn")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitOn indicates an expected call of WaitOn
func (mr *MockTriggerMockRecorder) WaitOn() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitOn", reflect.TypeOf((*MockTrigger)(nil).WaitOn))
}
