package botsfw

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
)

type BotRecordsFieldsSetter interface {

	// Platform returns platform name, e.g. 'telegram', 'fbmessenger', etc.
	// This method is for debug pruposes and to indicate that different platforms may have different fields
	// Though '*' can be used for a generic setter that works for all platforms
	// If both '*' and platform specific setters are defined, the generic setter will be used first.
	Platform() string

	// SetAppUserFields sets fields of app user record
	SetAppUserFields(appUser botsfwmodels.AppUserData, sender botinput.Sender) error

	// SetBotUserFields sets fields of bot user record
	SetBotUserFields(botUser botsfwmodels.PlatformUserData, sender botinput.Sender, botID, botUserID, appUserID string) error

	// SetBotChatFields sets fields of bot botChat record
	// TODO: document isAccessGranted parameter
	SetBotChatFields(botChat botsfwmodels.BotChatData, chat botinput.Chat, botID, botUserID, appUserID string, isAccessGranted bool) error
}
