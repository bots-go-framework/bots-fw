package bots

import "errors"

type AuthFailedError string

func (e AuthFailedError) Error() string {
	return string(e)
}

var (
	ErrEntityNotFound = errors.New("bots-framework: no such entity")
)