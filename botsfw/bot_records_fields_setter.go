package botsfw

import "github.com/bots-go-framework/bots-fw-store/botsfwmodels"

type BotRecordsFieldsSetter interface {

	// Platform returns platform name, e.g. 'telegram', 'fbmessenger', etc.
	// This method is for debug pruposes and to indicate that different platforms may have different fields
	// Though '*' can be used for a generic setter that works for all platforms
	// If both '*' and platform specific setters are defined, the generic setter will be used first.
	Platform() string

	// SetAppUserFields sets fields of app user record
	SetAppUserFields(appUser botsfwmodels.AppUserData, sender WebhookSender) error

	// SetBotUserFields sets fields of bot user record
	SetBotUserFields(botUser botsfwmodels.BotUserData, sender WebhookSender, botID, botUserID, appUserID string) error

	// SetBotChatFields sets fields of bot chat record
	// TODO: document isAccessGranted parameter
	SetBotChatFields(botChat botsfwmodels.ChatData, chat WebhookChat, botID, botUserID, appUserID string, isAccessGranted bool) error
}
