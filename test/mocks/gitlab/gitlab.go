// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.go
//
// Generated by this command:
//
//	mockgen -destination ../../test/mocks/gitlab/gitlab.go -source=gitlab.go
//

// Package mock_gitlab is a generated GoMock package.
package mock_gitlab

import (
	reflect "reflect"
	time "time"

	gitlab "github.com/iomarmochtar/gitlab-token-updater/pkg/gitlab"
	gomock "go.uber.org/mock/gomock"
)

// MockGitlabAPI is a mock of GitlabAPI interface.
type MockGitlabAPI struct {
	ctrl     *gomock.Controller
	recorder *MockGitlabAPIMockRecorder
	isgomock struct{}
}

// MockGitlabAPIMockRecorder is the mock recorder for MockGitlabAPI.
type MockGitlabAPIMockRecorder struct {
	mock *MockGitlabAPI
}

// NewMockGitlabAPI creates a new mock instance.
func NewMockGitlabAPI(ctrl *gomock.Controller) *MockGitlabAPI {
	mock := &MockGitlabAPI{ctrl: ctrl}
	mock.recorder = &MockGitlabAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitlabAPI) EXPECT() *MockGitlabAPIMockRecorder {
	return m.recorder
}

// Auth mocks base method.
func (m *MockGitlabAPI) Auth(token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Auth", token)
	ret0, _ := ret[0].(error)
	return ret0
}

// Auth indicates an expected call of Auth.
func (mr *MockGitlabAPIMockRecorder) Auth(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Auth", reflect.TypeOf((*MockGitlabAPI)(nil).Auth), token)
}

// GetGroupVar mocks base method.
func (m *MockGitlabAPI) GetGroupVar(path, varName string) (*gitlab.GitlabCICDVar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupVar", path, varName)
	ret0, _ := ret[0].(*gitlab.GitlabCICDVar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroupVar indicates an expected call of GetGroupVar.
func (mr *MockGitlabAPIMockRecorder) GetGroupVar(path, varName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupVar", reflect.TypeOf((*MockGitlabAPI)(nil).GetGroupVar), path, varName)
}

// GetRepoVar mocks base method.
func (m *MockGitlabAPI) GetRepoVar(path, varName string) (*gitlab.GitlabCICDVar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRepoVar", path, varName)
	ret0, _ := ret[0].(*gitlab.GitlabCICDVar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRepoVar indicates an expected call of GetRepoVar.
func (mr *MockGitlabAPIMockRecorder) GetRepoVar(path, varName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRepoVar", reflect.TypeOf((*MockGitlabAPI)(nil).GetRepoVar), path, varName)
}

// InitGitlab mocks base method.
func (m *MockGitlabAPI) InitGitlab(baseURL, token string) (gitlab.GitlabAPI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitGitlab", baseURL, token)
	ret0, _ := ret[0].(gitlab.GitlabAPI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InitGitlab indicates an expected call of InitGitlab.
func (mr *MockGitlabAPIMockRecorder) InitGitlab(baseURL, token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitGitlab", reflect.TypeOf((*MockGitlabAPI)(nil).InitGitlab), baseURL, token)
}

// ListGroupAccessToken mocks base method.
func (m *MockGitlabAPI) ListGroupAccessToken(path string) ([]gitlab.GitlabAccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGroupAccessToken", path)
	ret0, _ := ret[0].([]gitlab.GitlabAccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListGroupAccessToken indicates an expected call of ListGroupAccessToken.
func (mr *MockGitlabAPIMockRecorder) ListGroupAccessToken(path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGroupAccessToken", reflect.TypeOf((*MockGitlabAPI)(nil).ListGroupAccessToken), path)
}

// ListPersonalAccessToken mocks base method.
func (m *MockGitlabAPI) ListPersonalAccessToken() ([]gitlab.GitlabAccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPersonalAccessToken")
	ret0, _ := ret[0].([]gitlab.GitlabAccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPersonalAccessToken indicates an expected call of ListPersonalAccessToken.
func (mr *MockGitlabAPIMockRecorder) ListPersonalAccessToken() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPersonalAccessToken", reflect.TypeOf((*MockGitlabAPI)(nil).ListPersonalAccessToken))
}

// ListRepoAccessToken mocks base method.
func (m *MockGitlabAPI) ListRepoAccessToken(path string) ([]gitlab.GitlabAccessToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRepoAccessToken", path)
	ret0, _ := ret[0].([]gitlab.GitlabAccessToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRepoAccessToken indicates an expected call of ListRepoAccessToken.
func (mr *MockGitlabAPIMockRecorder) ListRepoAccessToken(path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRepoAccessToken", reflect.TypeOf((*MockGitlabAPI)(nil).ListRepoAccessToken), path)
}

// RotateGroupToken mocks base method.
func (m *MockGitlabAPI) RotateGroupToken(path string, tokenID int, expiredAt time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RotateGroupToken", path, tokenID, expiredAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RotateGroupToken indicates an expected call of RotateGroupToken.
func (mr *MockGitlabAPIMockRecorder) RotateGroupToken(path, tokenID, expiredAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RotateGroupToken", reflect.TypeOf((*MockGitlabAPI)(nil).RotateGroupToken), path, tokenID, expiredAt)
}

// RotatePersonalToken mocks base method.
func (m *MockGitlabAPI) RotatePersonalToken(tokenID int, expiredAt time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RotatePersonalToken", tokenID, expiredAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RotatePersonalToken indicates an expected call of RotatePersonalToken.
func (mr *MockGitlabAPIMockRecorder) RotatePersonalToken(tokenID, expiredAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RotatePersonalToken", reflect.TypeOf((*MockGitlabAPI)(nil).RotatePersonalToken), tokenID, expiredAt)
}

// RotateRepoToken mocks base method.
func (m *MockGitlabAPI) RotateRepoToken(path string, tokenID int, expiredAt time.Time) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RotateRepoToken", path, tokenID, expiredAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RotateRepoToken indicates an expected call of RotateRepoToken.
func (mr *MockGitlabAPIMockRecorder) RotateRepoToken(path, tokenID, expiredAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RotateRepoToken", reflect.TypeOf((*MockGitlabAPI)(nil).RotateRepoToken), path, tokenID, expiredAt)
}

// UpdateGroupVar mocks base method.
func (m *MockGitlabAPI) UpdateGroupVar(path, varName, value string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateGroupVar", path, varName, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateGroupVar indicates an expected call of UpdateGroupVar.
func (mr *MockGitlabAPIMockRecorder) UpdateGroupVar(path, varName, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateGroupVar", reflect.TypeOf((*MockGitlabAPI)(nil).UpdateGroupVar), path, varName, value)
}

// UpdateRepoVar mocks base method.
func (m *MockGitlabAPI) UpdateRepoVar(path, varName, value string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRepoVar", path, varName, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRepoVar indicates an expected call of UpdateRepoVar.
func (mr *MockGitlabAPIMockRecorder) UpdateRepoVar(path, varName, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRepoVar", reflect.TypeOf((*MockGitlabAPI)(nil).UpdateRepoVar), path, varName, value)
}
