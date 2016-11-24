package bots

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"fmt"
)

type BotEnvironment int8

const (
	EnvUnknown BotEnvironment = iota
	EnvProduction
	EnvStaging
	EnvDevTest
	EnvLocal
)

type BotSettings struct {
	Env         BotEnvironment
	Kind        string
	Code        string
	Token       string
	VerifyToken string // Used by Facebook
	Locale      strongo.Locale
}

func NewBotSettingsWithKind(env BotEnvironment, kind, code, token string, locale strongo.Locale) BotSettings {
	s := NewBotSettings(env, code, token, locale)
	s.Kind = kind
	return s
}

func NewBotSettings(mode BotEnvironment, code, token string, locale strongo.Locale) BotSettings {
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

type BotSettingsProvider func(c context.Context) BotSettingsBy

type BotSettingsBy struct { // TODO: Decide if it should have map[string]*BotSettings instead of map[string]BotSettings
	Code     map[string]BotSettings
	ApiToken map[string]BotSettings
	Locale   map[string][]BotSettings
}

func NewBotSettingsBy(bots ...BotSettings) BotSettingsBy {
	count := len(bots)
	botsBy := BotSettingsBy{
		Code:     make(map[string]BotSettings, count),
		ApiToken: make(map[string]BotSettings, count),
		Locale:   make(map[string][]BotSettings, count),
	}
	for _, bot := range bots {
		if _, ok := botsBy.Code[bot.Code]; ok {
			panic(fmt.Sprintf("Bot with duplicate code: %v", bot.Code))
		} else {
			botsBy.Code[bot.Code] = bot
		}
		if _, ok := botsBy.ApiToken[bot.Token]; ok {
			panic(fmt.Sprintf("Bot with duplicate token: %v", bot.Token))
		} else {
			botsBy.ApiToken[bot.Token] = bot
		}

		botsByLocale := botsBy.Locale[bot.Locale.Code5]
		botsByLocale = append(botsByLocale, bot)
		botsBy.Locale[bot.Locale.Code5] = botsByLocale
	}
	return botsBy
}
