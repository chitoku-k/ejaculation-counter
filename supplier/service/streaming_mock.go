// Code generated by MockGen. DO NOT EDIT.
// Source: streaming.go
//
// Generated by this command:
//
//	mockgen -source=streaming.go -destination=streaming_mock.go -package=service -self_package=github.com/chitoku-k/ejaculation-counter/supplier/service
//

// Package service is a generated GoMock package.
package service

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockStreaming is a mock of Streaming interface.
type MockStreaming struct {
	ctrl     *gomock.Controller
	recorder *MockStreamingMockRecorder
	isgomock struct{}
}

// MockStreamingMockRecorder is the mock recorder for MockStreaming.
type MockStreamingMockRecorder struct {
	mock *MockStreaming
}

// NewMockStreaming creates a new mock instance.
func NewMockStreaming(ctrl *gomock.Controller) *MockStreaming {
	mock := &MockStreaming{ctrl: ctrl}
	mock.recorder = &MockStreamingMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStreaming) EXPECT() *MockStreamingMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStreaming) Close(exit bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", exit)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStreamingMockRecorder) Close(exit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStreaming)(nil).Close), exit)
}

// Run mocks base method.
func (m *MockStreaming) Run(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockStreamingMockRecorder) Run(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockStreaming)(nil).Run), ctx)
}

// Statuses mocks base method.
func (m *MockStreaming) Statuses() <-chan Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Statuses")
	ret0, _ := ret[0].(<-chan Status)
	return ret0
}

// Statuses indicates an expected call of Statuses.
func (mr *MockStreamingMockRecorder) Statuses() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Statuses", reflect.TypeOf((*MockStreaming)(nil).Statuses))
}
