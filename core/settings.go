package bots

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"fmt"
)

type BotSettings struct {
	Env              strongo.Environment
	Kind             string
	Code             string
	Token            string
	PaymentToken     string
	PaymentTestToken string
	VerifyToken      string // Used by Facebook
	Locale           strongo.Locale
}

func NewBotSettingsWithKind(env strongo.Environment, kind, code, token string, locale strongo.Locale) BotSettings {
	s := NewBotSettings(env, code, token, locale)
	s.Kind = kind
	return s
}

func NewBotSettings(mode strongo.Environment, code, token string, locale strongo.Locale) BotSettings {
	if code == "" {
		panic("Missing required parameter: code")
	}
	if token == "" {
		panic("Missing required parameter: token")
	}
	if locale.Code5 == "" {
		panic("Missing required parameter: locale.Code5")
	}
	return BotSettings{
		Code:   code,
		Env:   mode,
		Token:  token,
		Locale: locale,
	}
}

type SettingsProvider func(c context.Context) SettingsBy

type SettingsBy struct {// TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
	Code     map[string]BotSettings
	ApiToken map[string]BotSettings
	Locale   map[string][]BotSettings
}

func NewBotSettingsBy(bots ...BotSettings) SettingsBy {
	count := len(bots)
	settingsBy := SettingsBy{
		Code:     make(map[string]BotSettings, count),
		ApiToken: make(map[string]BotSettings, count),
		Locale:   make(map[string][]BotSettings, count),
	}
	for _, bot := range bots {
		if _, ok := settingsBy.Code[bot.Code]; ok {
			panic(fmt.Sprintf("Bot with duplicate code: %v", bot.Code))
		} else {
			settingsBy.Code[bot.Code] = bot
		}
		if _, ok := settingsBy.ApiToken[bot.Token]; ok {
			panic(fmt.Sprintf("Bot with duplicate token: %v", bot.Token))
		} else {
			settingsBy.ApiToken[bot.Token] = bot
		}

		byLocale := settingsBy.Locale[bot.Locale.Code5]
		byLocale = append(byLocale, bot)
		settingsBy.Locale[bot.Locale.Code5] = byLocale
	}
	return settingsBy
}
