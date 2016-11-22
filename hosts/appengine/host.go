package gae_host

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"time"
)

type GaeBotHost struct {
}

var _ bots.BotHost = (*GaeBotHost)(nil)

func (h GaeBotHost) Logger(r *http.Request) strongo.Logger {
	return GaeLogger
}

func (h GaeBotHost) Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func (h GaeBotHost) GetHttpClient(r *http.Request) *http.Client {
	ctxWithDeadline, _ := context.WithTimeout(appengine.NewContext(r), 30*time.Second)
	return &http.Client{Transport: &urlfetch.Transport{Context: ctxWithDeadline}}
}

func (h GaeBotHost) GetBotCoreStores(platform string, appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	var (
		chatStore bots.BotChatStore
		userStore bots.BotUserStore
	)
	logger := h.Logger(r)
	appUserStore := NewGaeAppUserStore(logger, appContext.AppUserEntityKind(), appContext.AppUserEntityType(), appContext.NewBotAppUserEntity)

	switch platform { // TODO: Should not be hardcoded
	case "telegram":  // pass
		chatStore = NewGaeTelegramChatStore(logger)
		userStore = NewGaeTelegramUserStore(logger, appUserStore)
	case "fbm": 		// pass
		chatStore = NewGaeFbmChatStore(logger)
		userStore = NewGaeFacebookUserStore(logger, appUserStore)
	case "viber": 		// pass
		chatStore = NewGaeViberChatStore(logger)
		userStore = NewGaeViberUserStore(logger, appUserStore)
	default:
		panic("Unknown platform: " + platform)
	}

	return bots.BotCoreStores{
		BotChatStore:    chatStore,
		BotUserStore:    userStore,
		BotAppUserStore: appUserStore,
	}
}
