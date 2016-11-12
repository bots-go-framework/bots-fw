package fbm_strongo_bot

import (
	"encoding/json"
	"fmt"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"io/ioutil"
	"net/http"
	"strings"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"github.com/pkg/errors"
)

func NewFbmWebhookHandler(botsBy bots.BotSettingsBy, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) FbmWebhookHandler {
	if webhookDriver == nil {
		panic("webhookDriver == nil")
	}
	if botHost == nil {
		panic("botHost == nil")
	}
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return FbmWebhookHandler{
		BaseHandler: bots.BaseHandler{
			BotPlatform:        FbmPlatform{},
			BotHost:            botHost,
			WebhookDriver:      webhookDriver,
			TranslatorProvider: translatorProvider,
		},
		botsBy: botsBy,
	}
}

type FbmWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsBy
}

func (h FbmWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix+"/fbm/webhook", h.HandleWebhookRequest)
	http.HandleFunc(pathPrefix+"/fbm/webhook/", notFound) // TODO: Try to get rid?
	http.HandleFunc(pathPrefix+"/fbm/subscribe", h.Subscribe)
}

func (h FbmWebhookHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	httpClient := h.GetHttpClient(r)
	botCode := r.URL.Query().Get("bot")

	if botSettings, ok := h.botsBy.Code[botCode]; ok {
		res, err := httpClient.Post(fmt.Sprintf("https://graph.facebook.com/v2.6/me/subscribed_apps?access_token=%v", botSettings.Token), "", strings.NewReader(""))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error reading response body: %v", err)))
		} else {
			w.Write(body)
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (h FbmWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Debugf(c, "FbmWebhookHandler.HandleWebhookRequest()")
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		botCode := r.URL.Query().Get("bot")
		if botSettings, ok := h.botsBy.Code[botCode]; ok {
			var responseText string
			verifyToken := q.Get("hub.verify_token")
			if verifyToken == botSettings.VerifyToken {
				responseText = q.Get("hub.challenge")
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				responseText = "Wrong verify_token"
				log.Debugf(c, responseText + fmt.Sprintf(". Got: '%v', expected[bot=%v]: '%v'.", verifyToken, botCode, botSettings.VerifyToken))
			}
			w.Write([]byte(responseText))
		} else {
			log.Debugf(c, "Unkown bot '%v'", botCode)
			w.WriteHeader(http.StatusForbidden)
		}
	case http.MethodPost:
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h FbmWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	var receivedMessage fbm_bot_api.ReceivedMessage
	logger := h.BotHost.Logger(r)
	c := h.BotHost.Context(r)
	content := make([]byte, r.ContentLength)
	_, err = r.Body.Read(content)
	if err != nil {
		return
	}
	logger.Infof(c, "Request.Body: %v", string(content))
	err = json.Unmarshal(content, &receivedMessage)
	if err != nil {
		err = errors.Wrap(err, "Failed to deserialize FB json message")
		return
	}
	logger.Infof(c, "Unmarshaled JSON to a struct with %v entries: %v", len(receivedMessage.Entries), receivedMessage)
	entriesWithInputs = make([]bots.EntryInputs, len(receivedMessage.Entries))
	for i, entry := range receivedMessage.Entries {
		entryWithInputs := bots.EntryInputs{
			Entry:  entry,
			Inputs: make([]bots.WebhookInput, len(entry.Messaging)),
		}
		for j, messaging := range entry.Messaging {
			entryWithInputs.Inputs[j] = FbmWebhookInput{messaging: messaging}
		}
		entriesWithInputs[i] = entryWithInputs
	}
	botContext = &bots.BotContext{
		BotHost: h.BotHost,
		//BotSettings: nil, // TODO: fill with actual
	}
	return
}

func (h FbmWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewFbmWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (h FbmWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	panic("Not implemented yet") //return NewTelegramWebhookResponder(w, r)
}

func (h FbmWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(FbmPlatformID, appContext, r)
}
