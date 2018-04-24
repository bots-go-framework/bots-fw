package bots

import "errors"

// ErrNotImplemented if some feature is not implemented yet
var ErrNotImplemented = errors.New("Not implemented")

const (
	MessageTextBotDidNotUnderstandTheCommand = "MessageTextBotDidNotUnderstandTheCommand"
	MessageTextOopsSomethingWentWrong        = "MessageTextOopsSomethingWentWrong"
)
