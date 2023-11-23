// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pokt-network/poktroll/x/session/types (interfaces: QueryClient)

// Package mocksession is a generated GoMock package.
package mocksession

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/pokt-network/poktroll/x/session/types"
	grpc "google.golang.org/grpc"
)

// MockQueryClient is a mock of QueryClient interface.
type MockQueryClient struct {
	ctrl     *gomock.Controller
	recorder *MockQueryClientMockRecorder
}

// MockQueryClientMockRecorder is the mock recorder for MockQueryClient.
type MockQueryClientMockRecorder struct {
	mock *MockQueryClient
}

// NewMockQueryClient creates a new mock instance.
func NewMockQueryClient(ctrl *gomock.Controller) *MockQueryClient {
	mock := &MockQueryClient{ctrl: ctrl}
	mock.recorder = &MockQueryClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueryClient) EXPECT() *MockQueryClientMockRecorder {
	return m.recorder
}

// GetSession mocks base method.
func (m *MockQueryClient) GetSession(arg0 context.Context, arg1 *types.QueryGetSessionRequest, arg2 ...grpc.CallOption) (*types.QueryGetSessionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetSession", varargs...)
	ret0, _ := ret[0].(*types.QueryGetSessionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSession indicates an expected call of GetSession.
func (mr *MockQueryClientMockRecorder) GetSession(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSession", reflect.TypeOf((*MockQueryClient)(nil).GetSession), varargs...)
}

// Params mocks base method.
func (m *MockQueryClient) Params(arg0 context.Context, arg1 *types.QueryParamsRequest, arg2 ...grpc.CallOption) (*types.QueryParamsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Params", varargs...)
	ret0, _ := ret[0].(*types.QueryParamsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Params indicates an expected call of Params.
func (mr *MockQueryClientMockRecorder) Params(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Params", reflect.TypeOf((*MockQueryClient)(nil).Params), varargs...)
}
