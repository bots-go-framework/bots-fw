package gaehost

// GaeBaseStore is base store for GAE
type GaeBaseStore struct {
	// r *http.Request
	entityKind string
}

// NewGaeBaseStore creates base store for GAE
func NewGaeBaseStore(entityKind string) GaeBaseStore {
	return GaeBaseStore{entityKind: entityKind}
}
