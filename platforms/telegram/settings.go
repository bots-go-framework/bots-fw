package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app"
)

func NewTelegramBot(code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(code, token, locale)
}
