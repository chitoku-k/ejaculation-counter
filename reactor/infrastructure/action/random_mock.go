// Code generated by MockGen. DO NOT EDIT.
// Source: random.go
//
// Generated by this command:
//
//	mockgen -source=random.go -destination=random_mock.go -package=action -self_package=github.com/chitoku-k/ejaculation-counter/reactor/infrastructure/action
//

// Package action is a generated GoMock package.
package action

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockRandom is a mock of Random interface.
type MockRandom struct {
	ctrl     *gomock.Controller
	recorder *MockRandomMockRecorder
	isgomock struct{}
}

// MockRandomMockRecorder is the mock recorder for MockRandom.
type MockRandomMockRecorder struct {
	mock *MockRandom
}

// NewMockRandom creates a new mock instance.
func NewMockRandom(ctrl *gomock.Controller) *MockRandom {
	mock := &MockRandom{ctrl: ctrl}
	mock.recorder = &MockRandomMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRandom) EXPECT() *MockRandomMockRecorder {
	return m.recorder
}

// IntN mocks base method.
func (m *MockRandom) IntN(n int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IntN", n)
	ret0, _ := ret[0].(int)
	return ret0
}

// IntN indicates an expected call of IntN.
func (mr *MockRandomMockRecorder) IntN(n any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IntN", reflect.TypeOf((*MockRandom)(nil).IntN), n)
}
