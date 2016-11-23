package fbm_bot

import (
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
	"bytes"
	"github.com/pquerna/ffjson/ffjson"
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
var _ bots.WebhookHandler = (*FbmWebhookHandler)(nil)

func (handler FbmWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix+"/fbm/webhook", handler.HandleWebhookRequest)
	http.HandleFunc(pathPrefix+"/fbm/webhook/", notFound) // TODO: Try to get rid?
	http.HandleFunc(pathPrefix+"/fbm/subscribe", handler.Subscribe)
	http.HandleFunc(pathPrefix+"/fbm/whitelist", handler.Whitelist)
}

func (handler FbmWebhookHandler) Whitelist(w http.ResponseWriter, r *http.Request) {
	httpClient := handler.GetHttpClient(r)
	botCode := r.URL.Query().Get("bot")

	if botSettings, ok := handler.botsBy.Code[botCode]; ok {
		message := fbm_api.NewRequestWhitelistDomain("add", "https://" + r.URL.Host)
		requestBody, err := ffjson.MarshalFast(message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		log.Debugf(appengine.NewContext(r), "Posting to FB: %v", string(requestBody))
		res, err := httpClient.Post(fmt.Sprintf("https://graph.facebook.com/v2.6/me/thread_settings?access_token=%v", botSettings.Token), "application/json", bytes.NewReader(requestBody))
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

func (handler FbmWebhookHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	httpClient := handler.GetHttpClient(r)
	botCode := r.URL.Query().Get("bot")

	if botSettings, ok := handler.botsBy.Code[botCode]; ok {
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

func (handler FbmWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Debugf(c, "FbmWebhookHandler.HandleWebhookRequest()")
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		botCode := r.URL.Query().Get("bot")
		if botSettings, ok := handler.botsBy.Code[botCode]; ok {
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
		handler.HandleWebhook(w, r, handler)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (handler FbmWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	var receivedMessage fbm_api.ReceivedMessage
	logger := handler.BotHost.Logger(r)
	c := handler.BotHost.Context(r)
	content := make([]byte, r.ContentLength)
	_, err = r.Body.Read(content)
	if err != nil {
		return
	}
	logger.Infof(c, "Request.Body: %v", string(content))
	err = ffjson.UnmarshalFast(content, &receivedMessage)
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
			entryWithInputs.Inputs[j] = NewFbmWebhookInput(messaging)
		}
		entriesWithInputs[i] = entryWithInputs
	}

	botCode := r.URL.Query().Get("bot")
	if botSettings, ok := handler.botsBy.Code[botCode]; !ok {
		err = errors.New(fmt.Sprintf("Bot settings bot found by code: [%v]", botCode))
		return
	} else {
		botContext = bots.NewBotContext(handler.BotHost, botSettings);
	}
	return
}

func (_ FbmWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewFbmWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (_ FbmWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if fbmWhc, ok := whc.(*FbmWebhookContext); ok {
		return NewFbmWebhookResponder(fbmWhc)
	} else {
		panic(fmt.Sprintf("Expected FbmWebhookContext, got: %T", whc))
	}
}

func (handler FbmWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return handler.BotHost.GetBotCoreStores(FbmPlatformID, appContext, r)
}
