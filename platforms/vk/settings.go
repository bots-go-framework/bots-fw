package vk_bot

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func NewVkBot(env bots.BotEnvironment, kind, code, token string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettingsWithKind(env, kind, code, token, locale)
	botSettings.Locale = locale
	return botSettings
}
