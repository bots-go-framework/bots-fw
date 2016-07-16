package gae_host

import (
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
)

type GaeBotHost struct {
}

var _ bots.BotHost = (*GaeBotHost)(nil)

func (h GaeBotHost) GetLogger(r *http.Request) bots.Logger {
	return NewGaeLogger(appengine.NewContext(r))
}

func (h GaeBotHost) GetHttpClient(r *http.Request) *http.Client {
	return &http.Client{Transport: &urlfetch.Transport{Context: appengine.NewContext(r)}}
}

func (h GaeBotHost) GetBotCoreStores(platform string, appContext bots.AppContext, r *http.Request) bots.BotCoreStores {
	switch platform {
	case "telegram":
		logger := h.GetLogger(r)
		appUserStore := NewGaeAppUserStore(logger, r, appContext.AppUserEntityKind(), appContext.NewAppUserEntity)
		return bots.BotCoreStores{
			BotChatStore: NewGaeTelegramChatStore(logger, r),
			BotUserStore: NewGaeTelegramUserStore(logger, r, appUserStore),
			AppUserStore: appUserStore,
		}
	default:
		panic("Unknown platform: " + platform)
	}
}
