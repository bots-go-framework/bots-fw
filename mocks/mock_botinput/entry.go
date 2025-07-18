// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bots-go-framework/bots-fw/botinput (interfaces: Entry)
//
// Generated by this command:
//
//	mockgen github.com/bots-go-framework/bots-fw/botinput Entry
//

// Package mock_botinput is a generated GoMock package.
package mock_botinput

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockEntry is a mock of Entry interface.
type MockEntry struct {
	ctrl     *gomock.Controller
	recorder *MockEntryMockRecorder
	isgomock struct{}
}

// MockEntryMockRecorder is the mock recorder for MockEntry.
type MockEntryMockRecorder struct {
	mock *MockEntry
}

// NewMockEntry creates a new mock instance.
func NewMockEntry(ctrl *gomock.Controller) *MockEntry {
	mock := &MockEntry{ctrl: ctrl}
	mock.recorder = &MockEntryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEntry) EXPECT() *MockEntryMockRecorder {
	return m.recorder
}

// GetID mocks base method.
func (m *MockEntry) GetID() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(any)
	return ret0
}

// GetID indicates an expected call of GetID.
func (mr *MockEntryMockRecorder) GetID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockEntry)(nil).GetID))
}

// GetTime mocks base method.
func (m *MockEntry) GetTime() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTime")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetTime indicates an expected call of GetTime.
func (mr *MockEntryMockRecorder) GetTime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTime", reflect.TypeOf((*MockEntry)(nil).GetTime))
}
