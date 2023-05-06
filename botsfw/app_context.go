package botsfw

import (
	"github.com/strongo/app"
	"github.com/strongo/i18n"
)

// BotAppContext is a context for bot app
type BotAppContext interface {
	strongo.AppUserSettings // TODO: Do we really need it here?
	i18n.TranslationContext

	NewBotAppUserEntity() BotAppUser
	GetBotChatEntityFactory(platform string) func() BotChat
}
