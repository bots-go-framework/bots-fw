package gaehost

import (
	"context"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/dalgo/dal"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
)

// GaeBotHost represent information on current hosting platform
type GaeBotHost struct {
}

var _ bots.BotHost = (*GaeBotHost)(nil)

// Context creates context for http.Request
func (h GaeBotHost) Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

// GetHTTPClient creates an HTTP client using AppEngine's URL fetch
func (h GaeBotHost) GetHTTPClient(c context.Context) *http.Client {
	if c == nil {
		panic("c == nil")
	}
	return &http.Client{
		Transport: &urlfetch.Transport{
			Context: c,
		},
	}
}

// DB returns database instance
func (h GaeBotHost) DB() dal.Database {
	panic("not implemented")
	//return gaedb.NewDatabase()
}

// GetBotCoreStores returns bot DAL
func (h GaeBotHost) GetBotCoreStores(platform string, appContext bots.BotAppContext, r *http.Request) (stores bots.BotCoreStores) {
	appUserStore := NewGaeAppUserStore(appContext.AppUserEntityKind(), appContext.AppUserEntityType(), appContext.NewBotAppUserEntity)
	stores.BotAppUserStore = appUserStore

	switch platform { // TODO: Should not be hardcoded
	case "telegram": // pass
		panic("not implemented")
		//if tgChatStore := appContext.GetBotChatEntityFactory(platform); tgChatStore != nil {
		//	stores.BotChatStore = NewGaeTelegramChatStore(tgChatStore)
		//} else {
		//	stores.BotChatStore = NewGaeTelegramChatStore(func() bots.BotChat { return telegram.NewTelegramChatEntity() })
		//}
		//stores.BotUserStore = newGaeTelegramUserStore(appUserStore)
	case "fbm": // pass
		stores.BotChatStore = NewGaeFbmChatStore()
		stores.BotUserStore = newGaeFacebookUserStore(appUserStore)
	case "viber": // pass
		userChatStore := newGaeViberUserChatStore(appUserStore)
		stores.BotChatStore = userChatStore
		stores.BotUserStore = userChatStore
	default:
		panic("Unknown platform: " + platform)
	}
	return
}
