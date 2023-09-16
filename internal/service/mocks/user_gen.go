// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/user.go
//
// Generated by this command:
//
//	mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user_gen.go
//
// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	reflect "reflect"
	domain "webook/internal/domain"

	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// EditUserPassword mocks base method.
func (m *MockUserService) EditUserPassword(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EditUserPassword", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// EditUserPassword indicates an expected call of EditUserPassword.
func (mr *MockUserServiceMockRecorder) EditUserPassword(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EditUserPassword", reflect.TypeOf((*MockUserService)(nil).EditUserPassword), ctx, u)
}

// FindOrCreate mocks base method.
func (m *MockUserService) FindOrCreate(ctx context.Context, u domain.User) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreate", ctx, u)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreate indicates an expected call of FindOrCreate.
func (mr *MockUserServiceMockRecorder) FindOrCreate(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreate", reflect.TypeOf((*MockUserService)(nil).FindOrCreate), ctx, u)
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, u)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, u)
}

// Profile mocks base method.
func (m *MockUserService) Profile(ctx context.Context, u domain.User) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Profile", ctx, u)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Profile indicates an expected call of Profile.
func (mr *MockUserServiceMockRecorder) Profile(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Profile", reflect.TypeOf((*MockUserService)(nil).Profile), ctx, u)
}

// SignUp mocks base method.
func (m *MockUserService) SignUp(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignUp", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// SignUp indicates an expected call of SignUp.
func (mr *MockUserServiceMockRecorder) SignUp(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignUp", reflect.TypeOf((*MockUserService)(nil).SignUp), ctx, u)
}
