package vk

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

// NewVkBot defines VK bot
func NewVkBot(env strongo.Environment, profile, code, token string, locale strongo.Locale) bots.BotSettings {
	botSettings := bots.NewBotSettings(env, profile, code, "", token, locale)
	botSettings.Locale = locale
	return botSettings
}
