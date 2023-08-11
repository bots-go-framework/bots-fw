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
	Profile          BotProfile
	Code             string
	ID               string // TODO: Document how it is different from Code
	Token            string
	PaymentToken     string
	PaymentTestToken string
	VerifyToken      string // Used by Facebook
	GAToken          string // TODO: Refactor tu support multiple or move out

	// While a bot profile can support multiple locales a bot can be dedicated to a specific country/language
	Locale i18n.Locale
}

// NewBotSettings configures bot application
func NewBotSettings(platform Platform, mode strongo.Environment, profile BotProfile, code, id, token, gaToken string, locale i18n.Locale) BotSettings {
	if platform == "" {
		panic("NewBotSettings: missing required parameter: platform")
	}
	if profile == nil {
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
		panic("NewBotSettings: missing required parameter: Locale.Code5")
	}
	return BotSettings{
		Platform: platform,
		Profile:  profile,
		Code:     code,
		ID:       id,
		Env:      mode,
		Token:    token,
		GAToken:  gaToken,
		Locale:   locale,
	}
}

// SettingsProvider returns settings per different keys (ID, code, API token, Locale)
type SettingsProvider func(c context.Context) SettingsBy

// SettingsBy keeps settings per different keys (ID, code, API token, Locale)
// TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
type SettingsBy struct {

	// ByCode keeps settings by bot code - it is a human-readable ID of a bot
	ByCode map[string]*BotSettings

	// ByID keeps settings by bot ID - it is a machine-readable ID of a bot.
	ByID map[string]*BotSettings
}

// NewBotSettingsBy create settings per different keys (ID, code, API token, Locale)
func NewBotSettingsBy(bots ...BotSettings) (settingsBy SettingsBy) {
	count := len(bots)
	if count == 0 {
		panic("NewBotSettingsBy: missing required parameter: bots")
	}
	settingsBy = SettingsBy{
		ByCode: make(map[string]*BotSettings, count),
		ByID:   make(map[string]*BotSettings, count),
	}
	processBotSettings := func(i int, bot BotSettings) {
		if bot.Code == "" {
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
