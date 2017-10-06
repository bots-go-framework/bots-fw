package fbm_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewFbmBot(mode strongo.Environment, profile, code, id, token, verifyToken string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(mode, profile, code, id, token, locale)
	botSettings.VerifyToken = verifyToken
	return botSettings
}
