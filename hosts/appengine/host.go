package gae_host

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
)

type GaeBotHost struct {
}

var _ bots.BotHost = (*GaeBotHost)(nil)

func (h GaeBotHost) Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func (h GaeBotHost) GetHttpClient(c context.Context) *http.Client {
	return &http.Client{
		Transport: &urlfetch.Transport{
			Context: c,
		},
	}
}

func (h GaeBotHost) DB() db.Database {
	return gaedb.NewDatabase()
}


func (h GaeBotHost) GetBotCoreStores(platform string, appContext bots.BotAppContext, r *http.Request) (stores bots.BotCoreStores) {
	appUserStore := NewGaeAppUserStore(appContext.AppUserEntityKind(), appContext.AppUserEntityType(), appContext.NewBotAppUserEntity)
	stores.BotAppUserStore = appUserStore

	switch platform { // TODO: Should not be hardcoded
	case "telegram":  // pass
		stores.BotChatStore = NewGaeTelegramChatStore()
		stores.BotUserStore = NewGaeTelegramUserStore(appUserStore)
	case "fbm": 		// pass
		stores.BotChatStore = NewGaeFbmChatStore()
		stores.BotUserStore = NewGaeFacebookUserStore(appUserStore)
	case "viber": 		// pass
		userChatStore := NewGaeViberUserChatStore(appUserStore)
		stores.BotChatStore = userChatStore
		stores.BotUserStore = userChatStore
	default:
		panic("Unknown platform: " + platform)
	}
	return
}
