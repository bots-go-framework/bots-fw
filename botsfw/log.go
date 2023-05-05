package botsfw

import "context"
import logging "github.com/strongo/log"

type Logger interface {
	Infof(c context.Context, format string, args ...interface{})
	Errorf(c context.Context, format string, args ...interface{})
	Debugf(c context.Context, format string, args ...interface{})
	Warningf(c context.Context, format string, args ...interface{})
	Criticalf(c context.Context, format string, args ...interface{})
}

var log Logger = strongoLogger{}

type strongoLogger struct {
}

func (s strongoLogger) Infof(c context.Context, format string, args ...interface{}) {
	logging.Infof(c, format, args...)
}

func (s strongoLogger) Errorf(c context.Context, format string, args ...interface{}) {
	logging.Errorf(c, format, args...)
}

func (s strongoLogger) Debugf(c context.Context, format string, args ...interface{}) {
	logging.Debugf(c, format, args...)
}

func (s strongoLogger) Warningf(c context.Context, format string, args ...interface{}) {
	logging.Warningf(c, format, args...)
}

func (s strongoLogger) Criticalf(c context.Context, format string, args ...interface{}) {
	logging.Criticalf(c, format, args...)
}
