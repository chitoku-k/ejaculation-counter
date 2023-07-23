// Code generated by MockGen. DO NOT EDIT.
// Source: doublet.go

// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockDoubletRepository is a mock of DoubletRepository interface.
type MockDoubletRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDoubletRepositoryMockRecorder
}

// MockDoubletRepositoryMockRecorder is the mock recorder for MockDoubletRepository.
type MockDoubletRepositoryMockRecorder struct {
	mock *MockDoubletRepository
}

// NewMockDoubletRepository creates a new mock instance.
func NewMockDoubletRepository(ctrl *gomock.Controller) *MockDoubletRepository {
	mock := &MockDoubletRepository{ctrl: ctrl}
	mock.recorder = &MockDoubletRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDoubletRepository) EXPECT() *MockDoubletRepositoryMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockDoubletRepository) Get() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].([]string)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockDoubletRepositoryMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDoubletRepository)(nil).Get))
}
