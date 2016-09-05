package gae_host

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"net/http"
)

type GaeBaseStore struct {
	log        strongo.Logger
	r          *http.Request
	entityKind string
}

func (s *GaeBaseStore) Context() context.Context {
	return appengine.NewContext(s.r)
}

func NewGaeBaseStore(log strongo.Logger, r *http.Request, entityKind string) GaeBaseStore {
	return GaeBaseStore{log: log, r: r, entityKind: entityKind}
}
