package botsfw

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/i18n"
)

// BotAppContext is a context for bot app
type BotAppContext interface {
	//strongoapp.AppUserSettings // TODO: Do we really need it here?

	AppUserCollectionName() string

	i18n.TranslationContext

	NewBotAppUserEntity() botsfwmodels.AppUserData
	GetBotChatEntityFactory(platform string) func() botsfwmodels.BotChatData
}
