package bots

import "errors"

// ErrNotImplemented if some feature is not implemented yet
var ErrNotImplemented = errors.New("Not implemented")

const (
	MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND = "MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND"
	MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG        = "MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG"
)
