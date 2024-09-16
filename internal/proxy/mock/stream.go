// Code generated by MockGen. DO NOT EDIT.
// Source: cloud-proxy/proto/gen/proto/v1alpha (interfaces: CloudProxyAPI_StreamCloudProxyClient)

// Package mock_proxy is a generated GoMock package.
package mock_proxy

import (
	cloudproxyv1alpha "cloud-proxy/proto/gen/proto/v1alpha"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	metadata "google.golang.org/grpc/metadata"
)

// MockCloudProxyAPI_StreamCloudProxyClient is a mock of CloudProxyAPI_StreamCloudProxyClient interface.
type MockCloudProxyAPI_StreamCloudProxyClient struct {
	ctrl     *gomock.Controller
	recorder *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder
}

// MockCloudProxyAPI_StreamCloudProxyClientMockRecorder is the mock recorder for MockCloudProxyAPI_StreamCloudProxyClient.
type MockCloudProxyAPI_StreamCloudProxyClientMockRecorder struct {
	mock *MockCloudProxyAPI_StreamCloudProxyClient
}

// NewMockCloudProxyAPI_StreamCloudProxyClient creates a new mock instance.
func NewMockCloudProxyAPI_StreamCloudProxyClient(ctrl *gomock.Controller) *MockCloudProxyAPI_StreamCloudProxyClient {
	mock := &MockCloudProxyAPI_StreamCloudProxyClient{ctrl: ctrl}
	mock.recorder = &MockCloudProxyAPI_StreamCloudProxyClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) EXPECT() *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder {
	return m.recorder
}

// CloseSend mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) CloseSend() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseSend")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseSend indicates an expected call of CloseSend.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) CloseSend() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseSend", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).CloseSend))
}

// Context mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).Context))
}

// Header mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) Header() (metadata.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Header indicates an expected call of Header.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) Header() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).Header))
}

// Recv mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) Recv() (*cloudproxyv1alpha.StreamCloudProxyResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*cloudproxyv1alpha.StreamCloudProxyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).Recv))
}

// RecvMsg mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) Send(arg0 *cloudproxyv1alpha.StreamCloudProxyRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).Send), arg0)
}

// SendMsg mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).SendMsg), arg0)
}

// Trailer mocks base method.
func (m *MockCloudProxyAPI_StreamCloudProxyClient) Trailer() metadata.MD {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Trailer")
	ret0, _ := ret[0].(metadata.MD)
	return ret0
}

// Trailer indicates an expected call of Trailer.
func (mr *MockCloudProxyAPI_StreamCloudProxyClientMockRecorder) Trailer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trailer", reflect.TypeOf((*MockCloudProxyAPI_StreamCloudProxyClient)(nil).Trailer))
}
