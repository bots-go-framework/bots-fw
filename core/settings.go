package bots

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"fmt"
)

type BotSettings struct {
	Env              strongo.Environment
	ID               string
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

func NewBotSettingsWithID(env strongo.Environment, code, id, token string, locale strongo.Locale) BotSettings {
	s := NewBotSettings(env, code, token, locale)
	s.ID = id
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
		Env:    mode,
		Token:  token,
		Locale: locale,
	}
}

type SettingsProvider func(c context.Context) SettingsBy

type SettingsBy struct {
	// TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
	ByCode     map[string]BotSettings
	ByApiToken map[string]BotSettings
	ByLocale   map[string][]BotSettings
	ByID       map[string]BotSettings
}

func NewBotSettingsBy(bots ...BotSettings) SettingsBy {
	count := len(bots)
	settingsBy := SettingsBy{
		ByCode:     make(map[string]BotSettings, count),
		ByApiToken: make(map[string]BotSettings, count),
		ByLocale:   make(map[string][]BotSettings, count),
		ByID:       make(map[string]BotSettings, count),
	}
	for _, bot := range bots {
		if _, ok := settingsBy.ByCode[bot.Code]; ok {
			panic(fmt.Sprintf("Bot with duplicate code: %v", bot.Code))
		} else {
			settingsBy.ByCode[bot.Code] = bot
		}
		if _, ok := settingsBy.ByApiToken[bot.Token]; ok {
			panic(fmt.Sprintf("Bot with duplicate token: %v", bot.Token))
		} else {
			settingsBy.ByApiToken[bot.Token] = bot
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
