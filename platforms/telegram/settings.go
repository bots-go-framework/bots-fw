package telegram_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewTelegramBot(mode strongo.Environment, code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(mode, code, token, locale)
}
