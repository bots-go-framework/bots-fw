package telegram_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewTelegramBot(mode strongo.Environment, profile, code, token, paymentTestToken, paymentToken string, locale strongo.Locale) bots.BotSettings {
	settings := bots.NewBotSettings(mode, profile, code, "", token, locale)
	settings.PaymentTestToken = paymentTestToken
	settings.PaymentToken = paymentToken
	return settings
}
