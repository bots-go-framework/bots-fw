package gae_host

import (
	"golang.org/x/net/context"
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"github.com/strongo/bots-framework/core"
)

type GaeLogger struct {
	c context.Context
}

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

func NewGaeLogger(r *http.Request) GaeLogger {
	return GaeLogger{c: appengine.NewContext(r)}
}

type GaeBotHost struct {

}

func (h GaeBotHost) GetLogger(r *http.Request) bots.Logger {
	return NewGaeLogger(r)
}

func (h GaeBotHost) GetHttpClient(r *http.Request) *http.Client {
	return &http.Client{Transport: &urlfetch.Transport{Context: appengine.NewContext(r)}}
}
