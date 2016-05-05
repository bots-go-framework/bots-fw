package kik

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/bot"
)

const (
	KikChatKind = "KikChat"
	KikUserKind = "KikUser"
)

type KikUser struct {
	*bot.BotUserEntity
}

type KikChat struct {
	*bot.BotChatEntity
	KikUserID int
}
