package bots

import (
	"testing"
	"fmt"
)

type MockLogger struct {
	T        *testing.T
	Warnings []string
	Infos    []string
}

func (l *MockLogger) Debugf(format string, args ...interface{}) {
	l.T.Logf("DEBUG: " + format, args...)
}
func (l *MockLogger) Infof(format string, args ...interface{}) {
	l.Infos = append(l.Infos, fmt.Sprintf(format, args...))
	l.T.Logf("INFO: " + format, args...)
}
func (l *MockLogger) Warningf(format string, args ...interface{}) {
	l.T.Logf("WARNING: " + format, args...)
	l.Warnings = append(l.Warnings, fmt.Sprintf(format, args...))
}
func (l *MockLogger) Errorf(format string, args ...interface{}) {
	l.T.Logf("ERROR: " + format, args...)
}
func (l *MockLogger) Criticalf(format string, args ...interface{}) {
	l.T.Logf("CRITICAL: " + format, args...)
}

var _ Logger = (*MockLogger)(nil)
