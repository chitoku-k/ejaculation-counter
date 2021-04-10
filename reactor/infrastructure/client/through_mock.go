// Code generated by MockGen. DO NOT EDIT.
// Source: through.go

// Package client is a generated GoMock package.
package client

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockThrough is a mock of Through interface.
type MockThrough struct {
	ctrl     *gomock.Controller
	recorder *MockThroughMockRecorder
}

// MockThroughMockRecorder is the mock recorder for MockThrough.
type MockThroughMockRecorder struct {
	mock *MockThrough
}

// NewMockThrough creates a new mock instance.
func NewMockThrough(ctrl *gomock.Controller) *MockThrough {
	mock := &MockThrough{ctrl: ctrl}
	mock.recorder = &MockThroughMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockThrough) EXPECT() *MockThroughMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockThrough) Do(targetURL string) (ThroughResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", targetURL)
	ret0, _ := ret[0].(ThroughResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockThroughMockRecorder) Do(targetURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockThrough)(nil).Do), targetURL)
}