package fbm

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

// NewFbmBot creates definition of a new FB Messenger bot
func NewFbmBot(mode strongo.Environment, profile, code, id, token, verifyToken, gaToken string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(mode, profile, code, id, token, gaToken, locale)
	botSettings.VerifyToken = verifyToken
	return botSettings
}
