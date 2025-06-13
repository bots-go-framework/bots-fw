// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bots-go-framework/bots-fw/botinput (interfaces: WebhookAudioMessage)
//
// Generated by this command:
//
//	mockgen github.com/bots-go-framework/bots-fw/botinput WebhookAudioMessage
//

// Package mock_botinput is a generated GoMock package.
package mock_botinput

import (
	reflect "reflect"
	time "time"

	botinput "github.com/bots-go-framework/bots-fw/botinput"
	gomock "go.uber.org/mock/gomock"
)

// MockWebhookAudioMessage is a mock of WebhookAudioMessage interface.
type MockWebhookAudioMessage struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookAudioMessageMockRecorder
	isgomock struct{}
}

// MockWebhookAudioMessageMockRecorder is the mock recorder for MockWebhookAudioMessage.
type MockWebhookAudioMessageMockRecorder struct {
	mock *MockWebhookAudioMessage
}

// NewMockWebhookAudioMessage creates a new mock instance.
func NewMockWebhookAudioMessage(ctrl *gomock.Controller) *MockWebhookAudioMessage {
	mock := &MockWebhookAudioMessage{ctrl: ctrl}
	mock.recorder = &MockWebhookAudioMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebhookAudioMessage) EXPECT() *MockWebhookAudioMessageMockRecorder {
	return m.recorder
}

// BotChatID mocks base method.
func (m *MockWebhookAudioMessage) BotChatID() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BotChatID")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BotChatID indicates an expected call of BotChatID.
func (mr *MockWebhookAudioMessageMockRecorder) BotChatID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BotChatID", reflect.TypeOf((*MockWebhookAudioMessage)(nil).BotChatID))
}

// Chat mocks base method.
func (m *MockWebhookAudioMessage) Chat() botinput.WebhookChat {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Chat")
	ret0, _ := ret[0].(botinput.WebhookChat)
	return ret0
}

// Chat indicates an expected call of Chat.
func (mr *MockWebhookAudioMessageMockRecorder) Chat() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Chat", reflect.TypeOf((*MockWebhookAudioMessage)(nil).Chat))
}

// GetRecipient mocks base method.
func (m *MockWebhookAudioMessage) GetRecipient() botinput.WebhookRecipient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecipient")
	ret0, _ := ret[0].(botinput.WebhookRecipient)
	return ret0
}

// GetRecipient indicates an expected call of GetRecipient.
func (mr *MockWebhookAudioMessageMockRecorder) GetRecipient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecipient", reflect.TypeOf((*MockWebhookAudioMessage)(nil).GetRecipient))
}

// GetSender mocks base method.
func (m *MockWebhookAudioMessage) GetSender() botinput.WebhookUser {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSender")
	ret0, _ := ret[0].(botinput.WebhookUser)
	return ret0
}

// GetSender indicates an expected call of GetSender.
func (mr *MockWebhookAudioMessageMockRecorder) GetSender() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSender", reflect.TypeOf((*MockWebhookAudioMessage)(nil).GetSender))
}

// GetTime mocks base method.
func (m *MockWebhookAudioMessage) GetTime() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTime")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetTime indicates an expected call of GetTime.
func (mr *MockWebhookAudioMessageMockRecorder) GetTime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTime", reflect.TypeOf((*MockWebhookAudioMessage)(nil).GetTime))
}

// InputType mocks base method.
func (m *MockWebhookAudioMessage) InputType() botinput.WebhookInputType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InputType")
	ret0, _ := ret[0].(botinput.WebhookInputType)
	return ret0
}

// InputType indicates an expected call of InputType.
func (mr *MockWebhookAudioMessageMockRecorder) InputType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InputType", reflect.TypeOf((*MockWebhookAudioMessage)(nil).InputType))
}

// IntID mocks base method.
func (m *MockWebhookAudioMessage) IntID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IntID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// IntID indicates an expected call of IntID.
func (mr *MockWebhookAudioMessageMockRecorder) IntID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IntID", reflect.TypeOf((*MockWebhookAudioMessage)(nil).IntID))
}

// LogRequest mocks base method.
func (m *MockWebhookAudioMessage) LogRequest() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogRequest")
}

// LogRequest indicates an expected call of LogRequest.
func (mr *MockWebhookAudioMessageMockRecorder) LogRequest() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogRequest", reflect.TypeOf((*MockWebhookAudioMessage)(nil).LogRequest))
}

// StringID mocks base method.
func (m *MockWebhookAudioMessage) StringID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StringID")
	ret0, _ := ret[0].(string)
	return ret0
}

// StringID indicates an expected call of StringID.
func (mr *MockWebhookAudioMessageMockRecorder) StringID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StringID", reflect.TypeOf((*MockWebhookAudioMessage)(nil).StringID))
}
