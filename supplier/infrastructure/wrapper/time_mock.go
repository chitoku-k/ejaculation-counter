// Code generated by MockGen. DO NOT EDIT.
// Source: time.go

// Package wrapper is a generated GoMock package.
package wrapper

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockTicker is a mock of Ticker interface
type MockTicker struct {
	ctrl     *gomock.Controller
	recorder *MockTickerMockRecorder
}

// MockTickerMockRecorder is the mock recorder for MockTicker
type MockTickerMockRecorder struct {
	mock *MockTicker
}

// NewMockTicker creates a new mock instance
func NewMockTicker(ctrl *gomock.Controller) *MockTicker {
	mock := &MockTicker{ctrl: ctrl}
	mock.recorder = &MockTickerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTicker) EXPECT() *MockTickerMockRecorder {
	return m.recorder
}

// Tick mocks base method
func (m *MockTicker) Tick(d time.Duration) <-chan time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tick", d)
	ret0, _ := ret[0].(<-chan time.Time)
	return ret0
}

// Tick indicates an expected call of Tick
func (mr *MockTickerMockRecorder) Tick(d interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tick", reflect.TypeOf((*MockTicker)(nil).Tick), d)
}