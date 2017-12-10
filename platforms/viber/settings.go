package viber_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewViberBot(mode strongo.Environment, profile, code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(mode, profile, code, "", token, locale)
}
