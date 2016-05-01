package telegram_bot

import "github.com/strongo/bots-framework/core"

func NewTelegramBot(code, token string, locale bots.Locale) bots.BotSettings {
	return bots.NewBotSettings(code, token, locale)
}


