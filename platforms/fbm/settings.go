package fbm_strongo_bot

import "github.com/strongo/bots-framework/core"

func NewFbmBot(code, token, verifyToken string, locale bots.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(code, token, locale)
	botSettings.VerifyToken = verifyToken
	return botSettings
}