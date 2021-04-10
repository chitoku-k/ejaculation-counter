// Code generated by MockGen. DO NOT EDIT.
// Source: mpyw.go

// Package client is a generated GoMock package.
package client

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMpyw is a mock of Mpyw interface.
type MockMpyw struct {
	ctrl     *gomock.Controller
	recorder *MockMpywMockRecorder
}

// MockMpywMockRecorder is the mock recorder for MockMpyw.
type MockMpywMockRecorder struct {
	mock *MockMpyw
}

// NewMockMpyw creates a new mock instance.
func NewMockMpyw(ctrl *gomock.Controller) *MockMpyw {
	mock := &MockMpyw{ctrl: ctrl}
	mock.recorder = &MockMpywMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMpyw) EXPECT() *MockMpywMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockMpyw) Do(targetURL string, count int) (MpywChallengeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", targetURL, count)
	ret0, _ := ret[0].(MpywChallengeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockMpywMockRecorder) Do(targetURL, count interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockMpyw)(nil).Do), targetURL, count)
}