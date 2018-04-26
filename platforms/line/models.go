package line

import (
	"github.com/strongo/bots-framework/core"
)

const (
	// ChatKind return Line chat entity kind name
	ChatKind = "LineChat"

	// UserKind return Line user entity kind name
	UserKind = "LineUser"
)

// User is Line user entity
type User struct {
	*bots.BotUserEntity
}

// Chat is Line chat entity
type Chat struct {
	*bots.BotChatEntity
	LineUserID int
}
