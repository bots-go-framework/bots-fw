package gae_host

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"net/http"
)

type GaeBaseStore struct {
	logger     strongo.Logger
	r          *http.Request
	entityKind string
}

func (s *GaeBaseStore) Context() context.Context {
	return appengine.NewContext(s.r)
}

func NewGaeBaseStore(logger strongo.Logger, r *http.Request, entityKind string) GaeBaseStore {
	return GaeBaseStore{logger: logger, r: r, entityKind: entityKind}
}
