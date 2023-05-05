package botsfw

import (
	"context"
	"fmt"
	"testing"
)

var _ Logger = (*testLogger)(nil)

type testLogger struct {
	T        *testing.T
	Warnings []string
	Infos    []string
}

func (*testLogger) Name() string {
	return "testLogger"
}

func (l *testLogger) Debugf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("DEBUG: "+format, args...)
}
func (l *testLogger) Infof(c context.Context, format string, args ...interface{}) {
	l.Infos = append(l.Infos, fmt.Sprintf(format, args...))
	l.T.Logf("INFO: "+format, args...)
}
func (l *testLogger) Warningf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("WARNING: "+format, args...)
	l.Warnings = append(l.Warnings, fmt.Sprintf(format, args...))
}
func (l *testLogger) Errorf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("ERROR: "+format, args...)
}
func (l *testLogger) Criticalf(c context.Context, format string, args ...interface{}) {
	l.T.Logf("CRITICAL: "+format, args...)
}
