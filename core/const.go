package bots

import "errors"

var NotImplementedError = errors.New("Not implemented")

type InviteBy string

const (
	InviteByTelegram = InviteBy("telegram")
	InviteByFbm      = InviteBy("fbm")
	InviteByEmail    = InviteBy("email")
	InviteBySms      = InviteBy("sms")
)

const (
	MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND = "MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND"
	MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG        = "MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG"
)
