package gae_host

import (
	"net/http"
	"google.golang.org/appengine"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

type GaeBaseStore struct {
	log bots.Logger
	r *http.Request
	entityKind                string
}
func (s *GaeBaseStore) Context() context.Context {
	return appengine.NewContext(s.r)
}

func NewGaeBaseStore(log bots.Logger, r *http.Request, entityKind string) GaeBaseStore {
	return GaeBaseStore{log: log, r: r, entityKind: entityKind}
}