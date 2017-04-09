package gae_host

import (
	"golang.org/x/net/context"
	logGae "google.golang.org/appengine/log"
	"github.com/strongo/app/log"
)

type logger struct{}

func (l logger) Name() string {
	return ""
}
func (l logger) Debugf(c context.Context, format string, args ...interface{}) {
	if c == nil {
		panic("Required parameter 'c context.Context' is nill")
	}
	logGae.Debugf(c, format, args...)
}

func (l logger) Infof(c context.Context, format string, args ...interface{}) {
	if c == nil {
		panic("Required parameter 'c context.Context' is nill")
	}
	logGae.Infof(c, format, args...)
}

func (l logger) Warningf(c context.Context, format string, args ...interface{}) {
	if c == nil {
		panic("Required parameter 'c context.Context' is nill")
	}
	logGae.Warningf(c, format, args...)
}

func (l logger) Errorf(c context.Context, format string, args ...interface{}) {
	if c == nil {
		panic("Required parameter 'c context.Context' is nill")
	}
	logGae.Errorf(c, format, args...)
}

func (l logger) Criticalf(c context.Context, format string, args ...interface{}) {
	if c == nil {
		panic("Required parameter 'c context.Context' is nill")
	}
	logGae.Criticalf(c, format, args...)
}

var GaeLogger = (log.Logger)(logger{})
