package botsfw

import "github.com/bots-go-framework/bots-fw-store/botsfwmodels"

type BotRecordsFieldsSetter interface {
	Platform() string

	SetAppUserFields(appUser botsfwmodels.BotAppUser, sender WebhookSender) error

	SetBotUserFields(botUser botsfwmodels.BotUser, botID, botUserID, appUserID string, sender WebhookSender) error

	SetBotChatFields(botChat botsfwmodels.BotChat, botID, botUserID, appUserID string, chat WebhookChat, isAccessGranted bool) error
}
