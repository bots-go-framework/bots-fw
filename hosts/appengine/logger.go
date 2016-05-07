package gae_host

import (
	"github.com/strongo/bots-framework/core"
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"golang.org/x/net/context"
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

func NewGaeLogger(r *http.Request) GaeLogger {
	return GaeLogger{c: appengine.NewContext(r)}
}

