package line

import (
	"github.com/strongo/bots-framework/core"
)

const (
	LineChatKind = "LineChat"
	LineUserKind = "LineUser"
)

type LineUser struct {
	*bots.BotUserEntity
}

type LineChat struct {
	*bots.BotChatEntity
	LineUserID int
}
