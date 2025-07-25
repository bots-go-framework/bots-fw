// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bots-go-framework/bots-fw-store/botsfwmodels (interfaces: BotChatData)
//
// Generated by this command:
//
//	mockgen github.com/bots-go-framework/bots-fw-store/botsfwmodels BotChatData
//

// Package mock_botsfwmodels is a generated GoMock package.
package mock_botsfwmodels

import (
	reflect "reflect"
	time "time"

	botsfwmodels "github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	gomock "go.uber.org/mock/gomock"
)

// MockBotChatData is a mock of BotChatData interface.
type MockBotChatData struct {
	ctrl     *gomock.Controller
	recorder *MockBotChatDataMockRecorder
	isgomock struct{}
}

// MockBotChatDataMockRecorder is the mock recorder for MockBotChatData.
type MockBotChatDataMockRecorder struct {
	mock *MockBotChatData
}

// NewMockBotChatData creates a new mock instance.
func NewMockBotChatData(ctrl *gomock.Controller) *MockBotChatData {
	mock := &MockBotChatData{ctrl: ctrl}
	mock.recorder = &MockBotChatDataMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBotChatData) EXPECT() *MockBotChatDataMockRecorder {
	return m.recorder
}

// AddClientLanguage mocks base method.
func (m *MockBotChatData) AddClientLanguage(languageCode string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddClientLanguage", languageCode)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AddClientLanguage indicates an expected call of AddClientLanguage.
func (mr *MockBotChatDataMockRecorder) AddClientLanguage(languageCode any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddClientLanguage", reflect.TypeOf((*MockBotChatData)(nil).AddClientLanguage), languageCode)
}

// AddWizardParam mocks base method.
func (m *MockBotChatData) AddWizardParam(key, value string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddWizardParam", key, value)
}

// AddWizardParam indicates an expected call of AddWizardParam.
func (mr *MockBotChatDataMockRecorder) AddWizardParam(key, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWizardParam", reflect.TypeOf((*MockBotChatData)(nil).AddWizardParam), key, value)
}

// Base mocks base method.
func (m *MockBotChatData) Base() *botsfwmodels.ChatBaseData {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Base")
	ret0, _ := ret[0].(*botsfwmodels.ChatBaseData)
	return ret0
}

// Base indicates an expected call of Base.
func (mr *MockBotChatDataMockRecorder) Base() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Base", reflect.TypeOf((*MockBotChatData)(nil).Base))
}

// DelVar mocks base method.
func (m *MockBotChatData) DelVar(key string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DelVar", key)
}

// DelVar indicates an expected call of DelVar.
func (mr *MockBotChatDataMockRecorder) DelVar(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelVar", reflect.TypeOf((*MockBotChatData)(nil).DelVar), key)
}

// GetAppUserID mocks base method.
func (m *MockBotChatData) GetAppUserID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppUserID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAppUserID indicates an expected call of GetAppUserID.
func (mr *MockBotChatDataMockRecorder) GetAppUserID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppUserID", reflect.TypeOf((*MockBotChatData)(nil).GetAppUserID))
}

// GetAwaitingReplyTo mocks base method.
func (m *MockBotChatData) GetAwaitingReplyTo() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAwaitingReplyTo")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAwaitingReplyTo indicates an expected call of GetAwaitingReplyTo.
func (mr *MockBotChatDataMockRecorder) GetAwaitingReplyTo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAwaitingReplyTo", reflect.TypeOf((*MockBotChatData)(nil).GetAwaitingReplyTo))
}

// GetPreferredLanguage mocks base method.
func (m *MockBotChatData) GetPreferredLanguage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPreferredLanguage")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPreferredLanguage indicates an expected call of GetPreferredLanguage.
func (mr *MockBotChatDataMockRecorder) GetPreferredLanguage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPreferredLanguage", reflect.TypeOf((*MockBotChatData)(nil).GetPreferredLanguage))
}

// GetVar mocks base method.
func (m *MockBotChatData) GetVar(key string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVar", key)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetVar indicates an expected call of GetVar.
func (mr *MockBotChatDataMockRecorder) GetVar(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVar", reflect.TypeOf((*MockBotChatData)(nil).GetVar), key)
}

// GetWizardParam mocks base method.
func (m *MockBotChatData) GetWizardParam(key string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWizardParam", key)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetWizardParam indicates an expected call of GetWizardParam.
func (mr *MockBotChatDataMockRecorder) GetWizardParam(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWizardParam", reflect.TypeOf((*MockBotChatData)(nil).GetWizardParam), key)
}

// HasChangedVars mocks base method.
func (m *MockBotChatData) HasChangedVars() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasChangedVars")
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasChangedVars indicates an expected call of HasChangedVars.
func (mr *MockBotChatDataMockRecorder) HasChangedVars() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasChangedVars", reflect.TypeOf((*MockBotChatData)(nil).HasChangedVars))
}

// IsAccessGranted mocks base method.
func (m *MockBotChatData) IsAccessGranted() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAccessGranted")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAccessGranted indicates an expected call of IsAccessGranted.
func (mr *MockBotChatDataMockRecorder) IsAccessGranted() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAccessGranted", reflect.TypeOf((*MockBotChatData)(nil).IsAccessGranted))
}

// IsAwaitingReplyTo mocks base method.
func (m *MockBotChatData) IsAwaitingReplyTo(code string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAwaitingReplyTo", code)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAwaitingReplyTo indicates an expected call of IsAwaitingReplyTo.
func (mr *MockBotChatDataMockRecorder) IsAwaitingReplyTo(code any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAwaitingReplyTo", reflect.TypeOf((*MockBotChatData)(nil).IsAwaitingReplyTo), code)
}

// IsChanged mocks base method.
func (m *MockBotChatData) IsChanged() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsChanged")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsChanged indicates an expected call of IsChanged.
func (mr *MockBotChatDataMockRecorder) IsChanged() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsChanged", reflect.TypeOf((*MockBotChatData)(nil).IsChanged))
}

// IsGroupChat mocks base method.
func (m *MockBotChatData) IsGroupChat() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsGroupChat")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsGroupChat indicates an expected call of IsGroupChat.
func (mr *MockBotChatDataMockRecorder) IsGroupChat() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsGroupChat", reflect.TypeOf((*MockBotChatData)(nil).IsGroupChat))
}

// PopStepsFromAwaitingReplyUpToSpecificParent mocks base method.
func (m *MockBotChatData) PopStepsFromAwaitingReplyUpToSpecificParent(code string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PopStepsFromAwaitingReplyUpToSpecificParent", code)
}

// PopStepsFromAwaitingReplyUpToSpecificParent indicates an expected call of PopStepsFromAwaitingReplyUpToSpecificParent.
func (mr *MockBotChatDataMockRecorder) PopStepsFromAwaitingReplyUpToSpecificParent(code any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PopStepsFromAwaitingReplyUpToSpecificParent", reflect.TypeOf((*MockBotChatData)(nil).PopStepsFromAwaitingReplyUpToSpecificParent), code)
}

// PushStepToAwaitingReplyTo mocks base method.
func (m *MockBotChatData) PushStepToAwaitingReplyTo(code string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushStepToAwaitingReplyTo", code)
}

// PushStepToAwaitingReplyTo indicates an expected call of PushStepToAwaitingReplyTo.
func (mr *MockBotChatDataMockRecorder) PushStepToAwaitingReplyTo(code any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushStepToAwaitingReplyTo", reflect.TypeOf((*MockBotChatData)(nil).PushStepToAwaitingReplyTo), code)
}

// SetAccessGranted mocks base method.
func (m *MockBotChatData) SetAccessGranted(value bool) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAccessGranted", value)
	ret0, _ := ret[0].(bool)
	return ret0
}

// SetAccessGranted indicates an expected call of SetAccessGranted.
func (mr *MockBotChatDataMockRecorder) SetAccessGranted(value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccessGranted", reflect.TypeOf((*MockBotChatData)(nil).SetAccessGranted), value)
}

// SetAppUserID mocks base method.
func (m *MockBotChatData) SetAppUserID(appUserID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAppUserID", appUserID)
}

// SetAppUserID indicates an expected call of SetAppUserID.
func (mr *MockBotChatDataMockRecorder) SetAppUserID(appUserID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAppUserID", reflect.TypeOf((*MockBotChatData)(nil).SetAppUserID), appUserID)
}

// SetAwaitingReplyTo mocks base method.
func (m *MockBotChatData) SetAwaitingReplyTo(path string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAwaitingReplyTo", path)
}

// SetAwaitingReplyTo indicates an expected call of SetAwaitingReplyTo.
func (mr *MockBotChatDataMockRecorder) SetAwaitingReplyTo(path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAwaitingReplyTo", reflect.TypeOf((*MockBotChatData)(nil).SetAwaitingReplyTo), path)
}

// SetBotUserID mocks base method.
func (m *MockBotChatData) SetBotUserID(id any) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetBotUserID", id)
}

// SetBotUserID indicates an expected call of SetBotUserID.
func (mr *MockBotChatDataMockRecorder) SetBotUserID(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBotUserID", reflect.TypeOf((*MockBotChatData)(nil).SetBotUserID), id)
}

// SetDtLastInteraction mocks base method.
func (m *MockBotChatData) SetDtLastInteraction(time time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetDtLastInteraction", time)
}

// SetDtLastInteraction indicates an expected call of SetDtLastInteraction.
func (mr *MockBotChatDataMockRecorder) SetDtLastInteraction(time any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDtLastInteraction", reflect.TypeOf((*MockBotChatData)(nil).SetDtLastInteraction), time)
}

// SetIsGroupChat mocks base method.
func (m *MockBotChatData) SetIsGroupChat(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetIsGroupChat", arg0)
}

// SetIsGroupChat indicates an expected call of SetIsGroupChat.
func (mr *MockBotChatDataMockRecorder) SetIsGroupChat(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetIsGroupChat", reflect.TypeOf((*MockBotChatData)(nil).SetIsGroupChat), arg0)
}

// SetPreferredLanguage mocks base method.
func (m *MockBotChatData) SetPreferredLanguage(value string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPreferredLanguage", value)
}

// SetPreferredLanguage indicates an expected call of SetPreferredLanguage.
func (mr *MockBotChatDataMockRecorder) SetPreferredLanguage(value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPreferredLanguage", reflect.TypeOf((*MockBotChatData)(nil).SetPreferredLanguage), value)
}

// SetUpdatedTime mocks base method.
func (m *MockBotChatData) SetUpdatedTime(arg0 time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdatedTime", arg0)
}

// SetUpdatedTime indicates an expected call of SetUpdatedTime.
func (mr *MockBotChatDataMockRecorder) SetUpdatedTime(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdatedTime", reflect.TypeOf((*MockBotChatData)(nil).SetUpdatedTime), arg0)
}

// SetVar mocks base method.
func (m *MockBotChatData) SetVar(key, value string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetVar", key, value)
}

// SetVar indicates an expected call of SetVar.
func (mr *MockBotChatDataMockRecorder) SetVar(key, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVar", reflect.TypeOf((*MockBotChatData)(nil).SetVar), key, value)
}
