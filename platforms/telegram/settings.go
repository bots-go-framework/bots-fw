package telegram

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

// NewTelegramBot creates definition of new telegram bot
func NewTelegramBot(mode strongo.Environment, profile, code, token, paymentTestToken, paymentToken string, locale strongo.Locale) bots.BotSettings {
	settings := bots.NewBotSettings(mode, profile, code, "", token, locale)
	settings.PaymentTestToken = paymentTestToken
	settings.PaymentToken = paymentToken
	return settings
}
