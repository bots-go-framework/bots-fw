package botsfw

import (
	"context"
	"testing"
)

func TestInitBotsFrameworkLogger(t *testing.T) {
	InitBotsFrameworkLogger(testingLogger{t})
}

var _ Logger = testingLogger{}

// testingLogger is tests
type testingLogger struct {
	t *testing.T
}

// Debugf logs a debug message.
func (logger testingLogger) Debugf(_ context.Context, format string, args ...interface{}) {
	logger.t.Logf("Debug: "+format, args...)
}

// Infof is like Debugf, but at Info level.
func (logger testingLogger) Infof(_ context.Context, format string, args ...interface{}) {
	logger.t.Logf("Info: "+format, args...)
}

// Warningf is like Debugf, but at Warning level.
func (logger testingLogger) Warningf(_ context.Context, format string, args ...interface{}) {
	logger.t.Logf("Warning: "+format, args...)
}

// Errorf is like Debugf, but at Error level.
func (logger testingLogger) Errorf(_ context.Context, format string, args ...interface{}) {
	logger.t.Logf("Error: "+format, args...)
}

// Criticalf is like Debugf, but at Critical level.
func (logger testingLogger) Criticalf(_ context.Context, format string, args ...interface{}) {
	logger.t.Logf("Critical: "+format, args...)
}
