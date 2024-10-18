// Code generated by MockGen. DO NOT EDIT.
// Source: through.go
//
// Generated by this command:
//
//	mockgen -source=through.go -destination=through_mock.go -package=repository -self_package=github.com/chitoku-k/ejaculation-counter/reactor/repository
//

// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockThroughRepository is a mock of ThroughRepository interface.
type MockThroughRepository struct {
	ctrl     *gomock.Controller
	recorder *MockThroughRepositoryMockRecorder
	isgomock struct{}
}

// MockThroughRepositoryMockRecorder is the mock recorder for MockThroughRepository.
type MockThroughRepositoryMockRecorder struct {
	mock *MockThroughRepository
}

// NewMockThroughRepository creates a new mock instance.
func NewMockThroughRepository(ctrl *gomock.Controller) *MockThroughRepository {
	mock := &MockThroughRepository{ctrl: ctrl}
	mock.recorder = &MockThroughRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockThroughRepository) EXPECT() *MockThroughRepositoryMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockThroughRepository) Get() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockThroughRepositoryMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockThroughRepository)(nil).Get))
}
