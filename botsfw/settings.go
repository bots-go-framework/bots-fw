package botsfw

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app"
	"github.com/strongo/i18n"
	"os"
	"strings"
)

type DbGetter = func(ctx context.Context) (db dal.DB, err error)

type AppUserGetter = func(
	ctx context.Context,
	tx dal.ReadSession,
	botID string,
	appUserID string,
) (
	appUser record.DataWithID[string, botsfwmodels.AppUserData],
	err error,
)

// BotSettings keeps parameters of a bot that are static and are not changed in runtime
type BotSettings struct {

	// Platform is a platform that bot is running on
	// E.g.: Telegram, Viber, Facebook Messenger, WhatsApp, etc.
	Platform Platform

	// Env is an environment where bot is running
	// E.g.: Production/Live, Local/Dev, Staging, etc.
	Env strongo.Environment

	// Profile is a bot profile that defines bot's behavior
	// It includes commands router and some other settings
	// More in BotProfile documentation.
	Profile BotProfile

	// Code is a human-readable ID of a bot.
	// When displayed it is usually prefixed with @.
	// For example:
	//   - @listus_bot for https://t.me/listus_bot
	Code string

	// ID is a bot-platform ID of a bot. For example, it could be a GUID.
	// Not all platforms use it. For example Telegram doesn't use it.
	ID string

	// Token is used to authenticate bot with a platform when it is not responding to a webhook
	// but calling platform APIs directly.
	Token string

	// PaymentToken is used to process payments on bot platform
	PaymentToken string

	// PaymentTestToken is used to process test payments on bot platform
	PaymentTestToken string

	// VerifyToken is used by Facebook Messenger - TODO: Document how it is used and add a link to Facebook docs
	VerifyToken string

	// GAToken is Google Analytics token - TODO: Refactor tu support multiple or move out
	GAToken string

	// Locale is a default locale for a bot.
	// While a bot profile can support multiple locales a bot can be dedicated to a specific country/language
	Locale i18n.Locale

	// GetDatabase returns connection to a database assigned to a bot.
	// You can use same database for multiple bots
	// but if you need you can use different databases for different bots.
	// It's up to bots creator how to map bots to a database.
	// In most cases a single DB is used for all bots.
	GetDatabase DbGetter
	getAppUser  AppUserGetter
}

func (v BotSettings) GetAppUserByID(ctx context.Context, tx dal.ReadSession, appUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
	return v.getAppUser(ctx, tx, v.Code, appUserID)
}

// NewBotSettings configures bot application
func NewBotSettings(
	platform Platform,
	mode strongo.Environment,
	profile BotProfile,
	code, id, token, gaToken string,
	locale i18n.Locale,
	getDatabase DbGetter,
	getAppUser AppUserGetter,
) BotSettings {
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
		Platform:    platform,
		Profile:     profile,
		Code:        code,
		ID:          id,
		Env:         mode,
		Token:       token,
		GAToken:     gaToken,
		Locale:      locale,
		GetDatabase: getDatabase,
		getAppUser:  getAppUser,
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
