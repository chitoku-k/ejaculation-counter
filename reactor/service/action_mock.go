// Code generated by MockGen. DO NOT EDIT.
// Source: action.go
//
// Generated by this command:
//
//	mockgen -source=action.go -destination=action_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/reactor/service
//

// Package service is a generated GoMock package.
package service

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockAction is a mock of Action interface.
type MockAction struct {
	ctrl     *gomock.Controller
	recorder *MockActionMockRecorder
	isgomock struct{}
}

// MockActionMockRecorder is the mock recorder for MockAction.
type MockActionMockRecorder struct {
	mock *MockAction
}

// NewMockAction creates a new mock instance.
func NewMockAction(ctrl *gomock.Controller) *MockAction {
	mock := &MockAction{ctrl: ctrl}
	mock.recorder = &MockActionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAction) EXPECT() *MockActionMockRecorder {
	return m.recorder
}

// Event mocks base method.
func (m *MockAction) Event(ctx context.Context, message Message) (Event, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Event", ctx, message)
	ret0, _ := ret[0].(Event)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Event indicates an expected call of Event.
func (mr *MockActionMockRecorder) Event(ctx, message any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Event", reflect.TypeOf((*MockAction)(nil).Event), ctx, message)
}

// Name mocks base method.
func (m *MockAction) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockActionMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockAction)(nil).Name))
}

// Target mocks base method.
func (m *MockAction) Target(message Message) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Target", message)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Target indicates an expected call of Target.
func (mr *MockActionMockRecorder) Target(message any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Target", reflect.TypeOf((*MockAction)(nil).Target), message)
}
