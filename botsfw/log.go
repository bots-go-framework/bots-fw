package botsfw

import "context"

type Logger interface {
	Infof(c context.Context, format string, args ...interface{})
	Errorf(c context.Context, format string, args ...interface{})
	Debugf(c context.Context, format string, args ...interface{})
	Warningf(c context.Context, format string, args ...interface{})
	Criticalf(c context.Context, format string, args ...interface{})
}

var log Logger

func SetLogger(l Logger) {
	log = l
}

func Log() Logger {
	return log
}
