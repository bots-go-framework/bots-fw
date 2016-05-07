package gae_host

import "golang.org/x/net/context"

type GaeBaseStore struct {
	c context.Context
	entityKind                string
}