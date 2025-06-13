// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bots-go-framework/bots-fw/botinput (interfaces: WebhookConversationStarted)
//
// Generated by this command:
//
//	mockgen github.com/bots-go-framework/bots-fw/botinput WebhookConversationStarted
//

// Package mock_botinput is a generated GoMock package.
package mock_botinput

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockWebhookConversationStarted is a mock of WebhookConversationStarted interface.
type MockWebhookConversationStarted struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookConversationStartedMockRecorder
	isgomock struct{}
}

// MockWebhookConversationStartedMockRecorder is the mock recorder for MockWebhookConversationStarted.
type MockWebhookConversationStartedMockRecorder struct {
	mock *MockWebhookConversationStarted
}

// NewMockWebhookConversationStarted creates a new mock instance.
func NewMockWebhookConversationStarted(ctrl *gomock.Controller) *MockWebhookConversationStarted {
	mock := &MockWebhookConversationStarted{ctrl: ctrl}
	mock.recorder = &MockWebhookConversationStartedMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebhookConversationStarted) EXPECT() *MockWebhookConversationStartedMockRecorder {
	return m.recorder
}

// ConversationStartedMessage mocks base method.
func (m *MockWebhookConversationStarted) ConversationStartedMessage() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConversationStartedMessage")
	ret0, _ := ret[0].(any)
	return ret0
}

// ConversationStartedMessage indicates an expected call of ConversationStartedMessage.
func (mr *MockWebhookConversationStartedMockRecorder) ConversationStartedMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConversationStartedMessage", reflect.TypeOf((*MockWebhookConversationStarted)(nil).ConversationStartedMessage))
}
