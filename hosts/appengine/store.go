package gae_host

import (
	"github.com/strongo/app"
)

type GaeBaseStore struct {
	logger     strongo.Logger
	//r          *http.Request
	entityKind string
}

func NewGaeBaseStore(logger strongo.Logger, entityKind string) GaeBaseStore {
	return GaeBaseStore{logger: logger, entityKind: entityKind}
}
