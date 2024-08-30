package botsfw

import (
	"context"
	"errors"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
)

// BotContext binds a bot to a specific hosting environment
type BotContext struct {
	AppContext  AppContext
	BotHost     BotHost      // describes current bot app host environment
	BotSettings *BotSettings // keeps parameters of a bot that are static and are not changed in runtime
}

// BotContextProvider provides BotContext by platformID & botID
type BotContextProvider interface {
	// GetBotContext returns BotContext by platformID & botID
	GetBotContext(ctx context.Context, platformID botsfwconst.Platform, botID string) (botContext *BotContext, err error)
}

type botContextProvider struct {
	botHost            BotHost
	appContext         AppContext
	botSettingProvider BotSettingsProvider
}

func NewBotContextProvider(botHost BotHost, appContext AppContext, botSettingProvider BotSettingsProvider) BotContextProvider {
	if botHost == nil {
		panic("required argument botHost == nil")
	}
	if appContext == nil {
		panic("required argument appContext == nil")
	}
	if botSettingProvider == nil {
		panic("required argument botSettingProvider == nil")
	}
	return botContextProvider{
		appContext:         appContext,
		botHost:            botHost,
		botSettingProvider: botSettingProvider,
	}
}

var ErrUnknownBot = errors.New("unknown bot")

func (v botContextProvider) GetBotContext(ctx context.Context, platformID botsfwconst.Platform, botID string) (botContext *BotContext, err error) {
	botSettingsBy := v.botSettingProvider(ctx)
	botSettings, ok := botSettingsBy.ByID[botID]
	if !ok {
		if botSettings, ok = botSettingsBy.ByCode[botID]; !ok {
			return nil, ErrUnknownBot
		}
	}
	return &BotContext{
		AppContext:  v.appContext,
		BotHost:     v.botHost,
		BotSettings: botSettings,
	}, nil
}
