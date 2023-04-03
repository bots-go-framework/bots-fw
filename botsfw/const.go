package botsfw

import "errors"

// ErrNotImplemented if some feature is not implemented yet
var ErrNotImplemented = errors.New("not implemented")

const (
	// MessageTextBotDidNotUnderstandTheCommand is an i18n constant
	MessageTextBotDidNotUnderstandTheCommand = "MessageTextBotDidNotUnderstandTheCommand"

	// MessageTextOopsSomethingWentWrong is an i18n constant
	MessageTextOopsSomethingWentWrong = "MessageTextOopsSomethingWentWrong"
)
