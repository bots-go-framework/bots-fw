package botsfw

import (
	"context"
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/i18n"
	"os"
	"strings"
)

// BotSettings keeps parameters of a bot that are static and are not changed in runtime
type BotSettings struct {
	Platform         Platform
	Env              strongo.Environment
	ID               string
	Profile          string
	Code             string
	Token            string
	PaymentToken     string
	PaymentTestToken string
	VerifyToken      string // Used by Facebook
	Locale           i18n.Locale
	Router           WebhooksRouter
	GAToken          string // TODO: Refactor tu support multiple or move out
}

// NewBotSettings configures bot application
func NewBotSettings(platform Platform, mode strongo.Environment, profile, code, id, token, gaToken string, locale i18n.Locale) BotSettings {
	if platform == "" {
		panic("NewBotSettings: missing required parameter: platform")
	}
	if profile == "" {
		panic("NewBotSettings: missing required parameter: profile")
	}
	if code == "" {
		panic("NewBotSettings: missing required parameter: code")
	}
	if token == "" {
		envVarKey := fmt.Sprintf("%s_BOT_TOKEN_%s", strings.ToUpper(string(platform)), strings.ToUpper(code))
		token = os.Getenv(envVarKey)
		if token == "" {
			panic("NewBotSettings: missing required parameter 'token' and no environment variable " + envVarKey)
		}
	}
	if gaToken == "" {
		envVarKey := fmt.Sprintf("%s_GA_TOKEN_%s", strings.ToUpper(string(platform)), strings.ToUpper(code))
		gaToken = os.Getenv(envVarKey)
	}
	if locale.Code5 == "" {
		panic("NewBotSettings: missing required parameter: locale.Code5")
	}
	return BotSettings{
		Platform: platform,
		Profile:  profile,
		Code:     code,
		ID:       id,
		Env:      mode,
		Token:    token,
		Locale:   locale,
		GAToken:  gaToken,
	}
}

// SettingsProvider returns settings per different keys (ID, code, API token, locale)
type SettingsProvider func(c context.Context) SettingsBy

// SettingsBy keeps settings per different keys (ID, code, API token, locale)
// TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
type SettingsBy struct {

	// ByCode keeps settings by bot code - it is a human-readable ID of a bot
	ByCode map[string]*BotSettings

	// ByID keeps settings by bot ID - it is a machine-readable ID of a bot.
	ByID map[string]*BotSettings
}

// NewBotSettingsBy create settings per different keys (ID, code, API token, locale)
func NewBotSettingsBy(getRouter func(profile string) WebhooksRouter, bots ...BotSettings) (settingsBy SettingsBy) {
	count := len(bots)
	if count == 0 {
		panic("NewBotSettingsBy: missing required parameter: bots")
	}
	settingsBy = SettingsBy{
		ByCode: make(map[string]*BotSettings, count),
		ByID:   make(map[string]*BotSettings, count),
	}
	processBotSettings := func(i int, bot BotSettings) {
		if bot.Router.commandsByType == nil && getRouter != nil {
			bot.Router = getRouter(bot.Profile)
		}
		if bot.Code != "" {
			panic(fmt.Sprintf("Bot with empty code at index %v", i))
		}
		if _, ok := settingsBy.ByCode[bot.Code]; ok {
			panic(fmt.Sprintf("Bot with duplicate code: %v", bot.Code))
		} else {
			settingsBy.ByCode[bot.Code] = &bot
		}
		if bot.ID != "" {
			if _, ok := settingsBy.ByID[bot.ID]; ok {
				panic(fmt.Sprintf("Bot with duplicate ID: %v", bot.ID))
			} else {
				settingsBy.ByID[bot.ID] = &bot
			}
		}
	}
	for i, bot := range bots {
		processBotSettings(i, bot)
	}
	return settingsBy
}
