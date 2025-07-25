// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bots-go-framework/bots-fw/botinput (interfaces: Subscribed)
//
// Generated by this command:
//
//	mockgen github.com/bots-go-framework/bots-fw/botinput Subscribed
//

// Package mock_botinput is a generated GoMock package.
package mock_botinput

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockSubscribed is a mock of Subscribed interface.
type MockSubscribed struct {
	ctrl     *gomock.Controller
	recorder *MockSubscribedMockRecorder
	isgomock struct{}
}

// MockSubscribedMockRecorder is the mock recorder for MockSubscribed.
type MockSubscribedMockRecorder struct {
	mock *MockSubscribed
}

// NewMockSubscribed creates a new mock instance.
func NewMockSubscribed(ctrl *gomock.Controller) *MockSubscribed {
	mock := &MockSubscribed{ctrl: ctrl}
	mock.recorder = &MockSubscribedMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSubscribed) EXPECT() *MockSubscribedMockRecorder {
	return m.recorder
}

// SubscribedMessage mocks base method.
func (m *MockSubscribed) SubscribedMessage() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribedMessage")
	ret0, _ := ret[0].(any)
	return ret0
}

// SubscribedMessage indicates an expected call of SubscribedMessage.
func (mr *MockSubscribedMockRecorder) SubscribedMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribedMessage", reflect.TypeOf((*MockSubscribed)(nil).SubscribedMessage))
}
