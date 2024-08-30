package botsfw

import (
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/strongo/i18n"
)

// AppContext is a context for an app that uses the botsfw
type AppContext interface {
	//strongoapp.AppUserSettings // TODO: Do we really need it here?

	botsdal.AppUserDal
	//AppUserCollectionName() string
	//GetAppUserByBotUserID(ctx context.Context, platform, botID, botUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error)

	i18n.TranslationContext

	//NewBotAppUserEntity() botsfwmodels.AppUserData
	//GetBotChatEntityFactory(platform string) func() botsfwmodels.BotChatData
}
