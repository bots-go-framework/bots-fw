package kik

import "github.com/strongo/bots-framework/core"

const (
	// ChatKind is KIK chat entity kind name
	ChatKind = "KikChat"

	// UserKind is KIK user entity kind name
	UserKind = "KikUser"
)

// User is KIK user entity
type User struct {
	*bots.BotUserEntity
}

// Chat is KIK chat entity
type Chat struct {
	*bots.BotChatEntity
	KikUserID int
}
