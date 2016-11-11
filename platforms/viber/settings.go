package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app"
)


func NewViberBot(mode bots.BotMode, code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(mode, code, token, locale)
}
