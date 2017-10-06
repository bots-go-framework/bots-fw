package vk_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewVkBot(env strongo.Environment, profile, code, token string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(env, profile, code, "", token, locale)
	botSettings.Locale = locale
	return botSettings
}
