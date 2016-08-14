package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app"
)

func NewTelegramBot(mode bots.BotMode, code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(mode, code, token, locale)
}