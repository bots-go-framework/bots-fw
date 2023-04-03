package bots

import (
	"fmt"

	"context"
	"github.com/strongo/app"
)

// BotSettings keeps parameters of a bot
type BotSettings struct {
	Env              strongo.Environment
	ID               string
	Profile          string
	Code             string
	Token            string
	PaymentToken     string
	PaymentTestToken string
	VerifyToken      string // Used by Facebook
	Locale           strongo.Locale
	Router           WebhooksRouter
	GAToken          string // TODO: Refactor tu support multiple or move out
}

// NewBotSettings configures bot application
func NewBotSettings(mode strongo.Environment, profile, code, id, token, gaToken string, locale strongo.Locale) BotSettings {
	if profile == "" {
		panic("NewBotSettings: missing required parameter: profile")
	}
	if code == "" {
		panic("NewBotSettings: missing required parameter: code")
	}
	if token == "" {
		panic("NewBotSettings: missing required parameter: token")
	}
	if locale.Code5 == "" {
		panic("NewBotSettings: missing required parameter: locale.Code5")
	}
	return BotSettings{
		Profile: profile,
		Code:    code,
		ID:      id,
		Env:     mode,
		Token:   token,
		Locale:  locale,
		GAToken: gaToken,
	}
}

// SettingsProvider returns settings per different keys (ID, code, API token, locale)
type SettingsProvider func(c context.Context) SettingsBy

// SettingsBy keeps settings per different keys (ID, code, API token, locale)
type SettingsBy struct {
	// TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
	ByCode     map[string]BotSettings
	ByAPIToken map[string]BotSettings
	ByLocale   map[string][]BotSettings
	ByID       map[string]BotSettings
	HasRouter  bool
}

// NewBotSettingsBy create settings per different keys (ID, code, API token, locale)
func NewBotSettingsBy(router func(profile string) WebhooksRouter, bots ...BotSettings) SettingsBy {
	count := len(bots)
	settingsBy := SettingsBy{
		HasRouter:  router != nil,
		ByCode:     make(map[string]BotSettings, count),
		ByAPIToken: make(map[string]BotSettings, count),
		ByLocale:   make(map[string][]BotSettings, count),
		ByID:       make(map[string]BotSettings, count),
	}
	for _, bot := range bots {
		if settingsBy.HasRouter {
			bot.Router = router(bot.Profile)
		}
		if _, ok := settingsBy.ByCode[bot.Code]; ok {
			panic(fmt.Sprintf("Bot with duplicate code: %v", bot.Code))
		} else {
			settingsBy.ByCode[bot.Code] = bot
		}
		if _, ok := settingsBy.ByAPIToken[bot.Token]; ok {
			panic(fmt.Sprintf("Bot with duplicate token: %v", bot.Token))
		} else {
			settingsBy.ByAPIToken[bot.Token] = bot
		}
		if bot.ID != "" {
			if _, ok := settingsBy.ByID[bot.ID]; ok {
				panic(fmt.Sprintf("Bot with duplicate ID: %v", bot.ID))
			} else {
				settingsBy.ByID[bot.ID] = bot
			}
		}

		byLocale := settingsBy.ByLocale[bot.Locale.Code5]
		byLocale = append(byLocale, bot)
		settingsBy.ByLocale[bot.Locale.Code5] = byLocale

	}
	return settingsBy
}
