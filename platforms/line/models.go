package line

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/bot"
)

const (
	LineChatKind = "LineChat"
	LineUserKind = "LineUser"
)

type LineUser struct {
	*bot.BotUserEntity
}

type LineChat struct {
	*bot.BotChatEntity
	LineUserID int
}
