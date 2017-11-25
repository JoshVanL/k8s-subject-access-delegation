// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/subject_access_delegation/interfaces/interfaces.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	logrus "github.com/sirupsen/logrus"
	v1 "k8s.io/api/rbac/v1"
	informers "k8s.io/client-go/informers"
	kubernetes "k8s.io/client-go/kubernetes"
	reflect "reflect"
)

// MockSubjectAccessDelegation is a mock of SubjectAccessDelegation interface
type MockSubjectAccessDelegation struct {
	ctrl     *gomock.Controller
	recorder *MockSubjectAccessDelegationMockRecorder
}

// MockSubjectAccessDelegationMockRecorder is the mock recorder for MockSubjectAccessDelegation
type MockSubjectAccessDelegationMockRecorder struct {
	mock *MockSubjectAccessDelegation
}

// NewMockSubjectAccessDelegation creates a new mock instance
func NewMockSubjectAccessDelegation(ctrl *gomock.Controller) *MockSubjectAccessDelegation {
	mock := &MockSubjectAccessDelegation{ctrl: ctrl}
	mock.recorder = &MockSubjectAccessDelegationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubjectAccessDelegation) EXPECT() *MockSubjectAccessDelegationMockRecorder {
	return m.recorder
}

// Namespace mocks base method
func (m *MockSubjectAccessDelegation) Namespace() string {
	ret := m.ctrl.Call(m, "Namespace")
	ret0, _ := ret[0].(string)
	return ret0
}

// Namespace indicates an expected call of Namespace
func (mr *MockSubjectAccessDelegationMockRecorder) Namespace() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Namespace", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Namespace))
}

// Name mocks base method
func (m *MockSubjectAccessDelegation) Name() string {
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockSubjectAccessDelegationMockRecorder) Name() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Name))
}

// Kind mocks base method
func (m *MockSubjectAccessDelegation) Kind() string {
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind
func (mr *MockSubjectAccessDelegationMockRecorder) Kind() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Kind))
}

// Log mocks base method
func (m *MockSubjectAccessDelegation) Log() *logrus.Entry {
	ret := m.ctrl.Call(m, "Log")
	ret0, _ := ret[0].(*logrus.Entry)
	return ret0
}

// Log indicates an expected call of Log
func (mr *MockSubjectAccessDelegationMockRecorder) Log() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Log))
}

// Client mocks base method
func (m *MockSubjectAccessDelegation) Client() kubernetes.Interface {
	ret := m.ctrl.Call(m, "Client")
	ret0, _ := ret[0].(kubernetes.Interface)
	return ret0
}

// Client indicates an expected call of Client
func (mr *MockSubjectAccessDelegationMockRecorder) Client() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Client", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Client))
}

// KubeInformerFactory mocks base method
func (m *MockSubjectAccessDelegation) KubeInformerFactory() informers.SharedInformerFactory {
	ret := m.ctrl.Call(m, "KubeInformerFactory")
	ret0, _ := ret[0].(informers.SharedInformerFactory)
	return ret0
}

// KubeInformerFactory indicates an expected call of KubeInformerFactory
func (mr *MockSubjectAccessDelegationMockRecorder) KubeInformerFactory() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KubeInformerFactory", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).KubeInformerFactory))
}

// Triggers mocks base method
func (m *MockSubjectAccessDelegation) Triggers() []v1alpha1.EventTrigger {
	ret := m.ctrl.Call(m, "Triggers")
	ret0, _ := ret[0].([]v1alpha1.EventTrigger)
	return ret0
}

// Triggers indicates an expected call of Triggers
func (mr *MockSubjectAccessDelegationMockRecorder) Triggers() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Triggers", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Triggers))
}

// Delegate mocks base method
func (m *MockSubjectAccessDelegation) Delegate() (bool, error) {
	ret := m.ctrl.Call(m, "Delegate")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delegate indicates an expected call of Delegate
func (mr *MockSubjectAccessDelegationMockRecorder) Delegate() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delegate", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Delegate))
}

// DeleteRoleBindings mocks base method
func (m *MockSubjectAccessDelegation) DeleteRoleBindings() error {
	ret := m.ctrl.Call(m, "DeleteRoleBindings")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRoleBindings indicates an expected call of DeleteRoleBindings
func (mr *MockSubjectAccessDelegationMockRecorder) DeleteRoleBindings() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRoleBindings", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).DeleteRoleBindings))
}

// Delete mocks base method
func (m *MockSubjectAccessDelegation) Delete() error {
	ret := m.ctrl.Call(m, "Delete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockSubjectAccessDelegationMockRecorder) Delete() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).Delete))
}

// OriginSubject mocks base method
func (m *MockSubjectAccessDelegation) OriginSubject() OriginSubject {
	ret := m.ctrl.Call(m, "OriginSubject")
	ret0, _ := ret[0].(OriginSubject)
	return ret0
}

// OriginSubject indicates an expected call of OriginSubject
func (mr *MockSubjectAccessDelegationMockRecorder) OriginSubject() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OriginSubject", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).OriginSubject))
}

// DestinationSubjects mocks base method
func (m *MockSubjectAccessDelegation) DestinationSubjects() []DestinationSubject {
	ret := m.ctrl.Call(m, "DestinationSubjects")
	ret0, _ := ret[0].([]DestinationSubject)
	return ret0
}

// DestinationSubjects indicates an expected call of DestinationSubjects
func (mr *MockSubjectAccessDelegationMockRecorder) DestinationSubjects() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestinationSubjects", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).DestinationSubjects))
}

// ResolveDestinations mocks base method
func (m *MockSubjectAccessDelegation) ResolveDestinations() error {
	ret := m.ctrl.Call(m, "ResolveDestinations")
	ret0, _ := ret[0].(error)
	return ret0
}

// ResolveDestinations indicates an expected call of ResolveDestinations
func (mr *MockSubjectAccessDelegationMockRecorder) ResolveDestinations() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolveDestinations", reflect.TypeOf((*MockSubjectAccessDelegation)(nil).ResolveDestinations))
}

// MockOriginSubject is a mock of OriginSubject interface
type MockOriginSubject struct {
	ctrl     *gomock.Controller
	recorder *MockOriginSubjectMockRecorder
}

// MockOriginSubjectMockRecorder is the mock recorder for MockOriginSubject
type MockOriginSubjectMockRecorder struct {
	mock *MockOriginSubject
}

// NewMockOriginSubject creates a new mock instance
func NewMockOriginSubject(ctrl *gomock.Controller) *MockOriginSubject {
	mock := &MockOriginSubject{ctrl: ctrl}
	mock.recorder = &MockOriginSubjectMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOriginSubject) EXPECT() *MockOriginSubjectMockRecorder {
	return m.recorder
}

// ResolveOrigin mocks base method
func (m *MockOriginSubject) ResolveOrigin() error {
	ret := m.ctrl.Call(m, "ResolveOrigin")
	ret0, _ := ret[0].(error)
	return ret0
}

// ResolveOrigin indicates an expected call of ResolveOrigin
func (mr *MockOriginSubjectMockRecorder) ResolveOrigin() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolveOrigin", reflect.TypeOf((*MockOriginSubject)(nil).ResolveOrigin))
}

// RoleRefs mocks base method
func (m *MockOriginSubject) RoleRefs() ([]*v1.RoleRef, error) {
	ret := m.ctrl.Call(m, "RoleRefs")
	ret0, _ := ret[0].([]*v1.RoleRef)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RoleRefs indicates an expected call of RoleRefs
func (mr *MockOriginSubjectMockRecorder) RoleRefs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RoleRefs", reflect.TypeOf((*MockOriginSubject)(nil).RoleRefs))
}

// Name mocks base method
func (m *MockOriginSubject) Name() string {
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockOriginSubjectMockRecorder) Name() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockOriginSubject)(nil).Name))
}

// Kind mocks base method
func (m *MockOriginSubject) Kind() string {
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind
func (mr *MockOriginSubjectMockRecorder) Kind() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockOriginSubject)(nil).Kind))
}

// MockDestinationSubject is a mock of DestinationSubject interface
type MockDestinationSubject struct {
	ctrl     *gomock.Controller
	recorder *MockDestinationSubjectMockRecorder
}

// MockDestinationSubjectMockRecorder is the mock recorder for MockDestinationSubject
type MockDestinationSubjectMockRecorder struct {
	mock *MockDestinationSubject
}

// NewMockDestinationSubject creates a new mock instance
func NewMockDestinationSubject(ctrl *gomock.Controller) *MockDestinationSubject {
	mock := &MockDestinationSubject{ctrl: ctrl}
	mock.recorder = &MockDestinationSubjectMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDestinationSubject) EXPECT() *MockDestinationSubjectMockRecorder {
	return m.recorder
}

// ResolveDestination mocks base method
func (m *MockDestinationSubject) ResolveDestination() error {
	ret := m.ctrl.Call(m, "ResolveDestination")
	ret0, _ := ret[0].(error)
	return ret0
}

// ResolveDestination indicates an expected call of ResolveDestination
func (mr *MockDestinationSubjectMockRecorder) ResolveDestination() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolveDestination", reflect.TypeOf((*MockDestinationSubject)(nil).ResolveDestination))
}

// Name mocks base method
func (m *MockDestinationSubject) Name() string {
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockDestinationSubjectMockRecorder) Name() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockDestinationSubject)(nil).Name))
}

// Kind mocks base method
func (m *MockDestinationSubject) Kind() string {
	ret := m.ctrl.Call(m, "Kind")
	ret0, _ := ret[0].(string)
	return ret0
}

// Kind indicates an expected call of Kind
func (mr *MockDestinationSubjectMockRecorder) Kind() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kind", reflect.TypeOf((*MockDestinationSubject)(nil).Kind))
}

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
