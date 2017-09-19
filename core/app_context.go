package bots

import "github.com/strongo/app"

type BotAppContext interface {
	strongo.AppContext
	NewBotAppUserEntity() BotAppUser
	GetBotChatEntityFactory(platform string) func() BotChat
}
