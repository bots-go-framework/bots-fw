package kik

import ()
import "github.com/strongo/bots-framework/core"

const (
	KikChatKind = "KikChat"
	KikUserKind = "KikUser"
)

type KikUser struct {
	*bots.BotUserEntity
}

type KikChat struct {
	*bots.BotChatEntity
	KikUserID int
}
