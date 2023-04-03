package bots

import "github.com/strongo/app"

// BotAppContext is a context for bot app
type BotAppContext interface {
	strongo.AppContext
	NewBotAppUserEntity() BotAppUser
	GetBotChatEntityFactory(platform string) func() BotChat
}
