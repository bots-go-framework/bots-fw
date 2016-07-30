package fbm_strongo_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app"
)

func NewFbmBot(code, token, verifyToken string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(code, token, locale)
	botSettings.VerifyToken = verifyToken
	return botSettings
}
