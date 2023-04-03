package botsfw

import "errors"

// ErrAuthFailed raised if authentication failed
type ErrAuthFailed string

func (e ErrAuthFailed) Error() string {
	return string(e)
}

var (
	// ErrEntityNotFound is returned if entity not found in storage
	ErrEntityNotFound = errors.New("bots-framework: no such entity")
)
