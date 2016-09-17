package gae_host

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type logger struct{}

var _ strongo.Logger = (*logger)(nil)

func (l logger) Debugf(c context.Context, format string, args ...interface{}) {
	log.Debugf(c, format, args...)
}

func (l logger) Infof(c context.Context, format string, args ...interface{}) {
	log.Infof(c, format, args...)
}

func (l logger) Warningf(c context.Context, format string, args ...interface{}) {
	log.Warningf(c, format, args...)
}

func (l logger) Errorf(c context.Context, format string, args ...interface{}) {
	log.Errorf(c, format, args...)
}

func (l logger) Criticalf(c context.Context, format string, args ...interface{}) {
	log.Criticalf(c, format, args...)
}

var GaeLogger = (strongo.Logger)(logger{})
