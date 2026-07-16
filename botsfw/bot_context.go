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

// GetBotContext returns the BotContext for a bot on a specific platform.
//
// The lookup is scoped to platformID. This matters: before it was, this method
// accepted platformID and ignored it, resolving purely by bot ID or code. With a
// single platform that was invisible. With two, a webhook for platform A could
// resolve platform B's settings — including its Token and WebhookSecretToken —
// whenever the two shared an ID or code.
//
// Returns ErrUnknownBot when no bot with that ID or code exists ON THAT PLATFORM,
// even if one exists elsewhere. That is deliberate: a caller asking for a WhatsApp
// bot must never be handed a Telegram one.
func (v botContextProvider) GetBotContext(ctx context.Context, platformID botsfwconst.Platform, botID string) (botContext *BotContext, err error) {
	if platformID == "" {
		return nil, errors.New("required parameter platformID is empty")
	}
	botSettingsBy := v.botSettingProvider(ctx)

	botSettings := botSettingsBy.findByPlatform(platformID, botID)
	if botSettings == nil {
		return nil, ErrUnknownBot
	}
	return &BotContext{
		AppContext:  v.appContext,
		BotHost:     v.botHost,
		BotSettings: botSettings,
	}, nil
}
