package fbm_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewFbmBot(mode strongo.Environment, code, id, token, verifyToken string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettingsWithID(mode, code, id, token, locale)
	botSettings.VerifyToken = verifyToken
	botSettings.Locale = locale
	return botSettings
}
