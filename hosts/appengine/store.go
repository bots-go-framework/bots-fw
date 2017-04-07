package gae_host

type GaeBaseStore struct {
	//r          *http.Request
	entityKind string
}

func NewGaeBaseStore(entityKind string) GaeBaseStore {
	return GaeBaseStore{entityKind: entityKind}
}
