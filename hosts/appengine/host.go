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
	return NewGaeLogger(r)
}

func (h GaeBotHost) GetHttpClient(r *http.Request) *http.Client {
	return &http.Client{Transport: &urlfetch.Transport{Context: appengine.NewContext(r)}}
}

func (h GaeBotHost) GetBotChatStore(platform string, r *http.Request) bots.BotChatStore {
	switch platform {
	case "telegram": return NewGaeTelegramChatStore(appengine.NewContext(r))
	default: panic("Unknown platform: " + platform)
	}
}
