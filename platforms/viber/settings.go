package viber

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

// NewViberBot creates definition of Viber bot
func NewViberBot(mode strongo.Environment, profile, code, token string, locale strongo.Locale) bots.BotSettings {
	return bots.NewBotSettings(mode, profile, code, "", token, locale)
}
