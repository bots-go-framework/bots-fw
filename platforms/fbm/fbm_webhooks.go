package fbm_bot

import (
	"fmt"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"github.com/pkg/errors"
	"bytes"
	"github.com/pquerna/ffjson/ffjson"
	"golang.org/x/net/context"
	"github.com/julienschmidt/httprouter"
)

func NewFbmWebhookHandler(botsBy bots.SettingsProvider, translatorProvider bots.TranslatorProvider) FbmWebhookHandler {
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return FbmWebhookHandler{
		BaseHandler: bots.BaseHandler{
			BotPlatform:        FbmPlatform{},
			TranslatorProvider: translatorProvider,
		},
		bots: botsBy,
	}
}

type FbmWebhookHandler struct {
	bots.BaseHandler
	bots bots.SettingsProvider
}
var _ bots.WebhookHandler = (*FbmWebhookHandler)(nil)

func (handler FbmWebhookHandler) RegisterWebhookHandler(driver bots.WebhookDriver, host bots.BotHost, router *httprouter.Router, pathPrefix string) {
	if router == nil {
		panic("router == nil")
	}
	handler.BaseHandler.Register(driver, host)
	router.POST(pathPrefix+"/fbm/webhook", handler.HandleWebhookRequest)
	router.POST(pathPrefix+"/fbm/subscribe", handler.Subscribe)
	router.POST(pathPrefix+"/fbm/whitelist", handler.Whitelist)
}


func (handler FbmWebhookHandler) Whitelist(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := handler.Context(r)
	httpClient := handler.GetHttpClient(c)
	botCode := r.URL.Query().Get("bot")

	fbmBots := handler.bots(c)
	if botSettings, ok := fbmBots.ByCode[botCode]; ok {
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

func (handler FbmWebhookHandler) Subscribe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := handler.Context(r)
	httpClient := handler.GetHttpClient(c)
	botCode := r.URL.Query().Get("bot")

	fbmBots := handler.bots(c)
	if botSettings, ok := fbmBots.ByCode[botCode]; ok {
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

func (handler FbmWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	log.Debugf(c, "FbmWebhookHandler.HandleWebhookRequest()")
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		botCode := r.URL.Query().Get("bot")
		fbmBots := handler.bots(c)
		if botSettings, ok := fbmBots.ByCode[botCode]; ok {
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

func (handler FbmWebhookHandler) GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	var (
		receivedMessage fbm_api.ReceivedMessage
		bodyBytes []byte
	)
	defer r.Body.Close()
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		errors.Wrap(err, "Failed to read request body")
		return
	}
	log.Infof(c, "Request.Body: %v", string(bodyBytes))
	err = ffjson.UnmarshalFast(bodyBytes, &receivedMessage)
	if err != nil {
		err = errors.Wrap(err, "Failed to deserialize FB json message")
		return
	}
	log.Infof(c, "Unmarshaled JSON to a struct with %v entries: %v", len(receivedMessage.Entries), receivedMessage)
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

	//botCode := r.URL.Query().Get("bot")
	pageID := receivedMessage.Entries[0].Messaging[0].Recipient.ID
	fbmBots := handler.bots(c)
	if botSettings, ok := fbmBots.ByID[pageID]; !ok {
		err = fmt.Errorf("Bot settings not found by ID: [%v]", pageID)
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
