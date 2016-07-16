package gae_host

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type GaeLogger struct {
	c context.Context
}

var _ bots.Logger = (*GaeLogger)(nil)

func (l GaeLogger) Debugf(format string, args ...interface{}) {
	log.Debugf(l.c, format, args...)
}

func (l GaeLogger) Infof(format string, args ...interface{}) {
	log.Infof(l.c, format, args...)
}

func (l GaeLogger) Warningf(format string, args ...interface{}) {
	log.Warningf(l.c, format, args...)
}
func (l GaeLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(l.c, format, args...)
}
func (l GaeLogger) Criticalf(format string, args ...interface{}) {
	log.Criticalf(l.c, format, args...)
}

func NewGaeLogger(c context.Context) GaeLogger {
	return GaeLogger{c: c}
}
