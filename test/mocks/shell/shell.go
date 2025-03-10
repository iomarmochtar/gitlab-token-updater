// Code generated by MockGen. DO NOT EDIT.
// Source: shell.go
//
// Generated by this command:
//
//	mockgen -destination ../../test/mocks/shell/shell.go -source=shell.go
//

// Package mock_shell is a generated GoMock package.
package mock_shell

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockShell is a mock of Shell interface.
type MockShell struct {
	ctrl     *gomock.Controller
	recorder *MockShellMockRecorder
	isgomock struct{}
}

// MockShellMockRecorder is the mock recorder for MockShell.
type MockShellMockRecorder struct {
	mock *MockShell
}

// NewMockShell creates a new mock instance.
func NewMockShell(ctrl *gomock.Controller) *MockShell {
	mock := &MockShell{ctrl: ctrl}
	mock.recorder = &MockShellMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockShell) EXPECT() *MockShellMockRecorder {
	return m.recorder
}

// Exec mocks base method.
func (m *MockShell) Exec(command string, envVars map[string]string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exec", command, envVars)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockShellMockRecorder) Exec(command, envVars any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockShell)(nil).Exec), command, envVars)
}

// FileMustExists mocks base method.
func (m *MockShell) FileMustExists(path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FileMustExists", path)
	ret0, _ := ret[0].(error)
	return ret0
}

// FileMustExists indicates an expected call of FileMustExists.
func (mr *MockShellMockRecorder) FileMustExists(path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FileMustExists", reflect.TypeOf((*MockShell)(nil).FileMustExists), path)
}
