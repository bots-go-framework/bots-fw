package bots

import (
	"fmt"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"testing"
)

type MockLogger struct {
	T        *testing.T
	Warnings []string
	Infos    []string
}

func (_ *MockLogger) Name() string {
	return "MockLogger"
}

func (l *MockLogger) Debugf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("DEBUG: "+format, args...)
}
func (l *MockLogger) Infof(c context.Context, format string, args ...interface{}) {
	l.Infos = append(l.Infos, fmt.Sprintf(format, args...))
	l.T.Logf("INFO: "+format, args...)
}
func (l *MockLogger) Warningf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("WARNING: "+format, args...)
	l.Warnings = append(l.Warnings, fmt.Sprintf(format, args...))
}
func (l *MockLogger) Errorf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("ERROR: "+format, args...)
}
func (l *MockLogger) Criticalf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("CRITICAL: "+format, args...)
}

var _ log.Logger = (*MockLogger)(nil)
