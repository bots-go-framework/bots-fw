package fbm_strongo_bot

import (
	"encoding/json"
	"fmt"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewFbmWebhookHandler(botsBy bots.BotSettingsBy, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) FbmWebhookHandler {
	return FbmWebhookHandler{
		BaseHandler: bots.BaseHandler{
			BotPlatform:   FbmPlatform{},
			BotHost:       botHost,
			WebhookDriver: webhookDriver,
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
	http.HandleFunc(pathPrefix + "/fbm/webhook", h.HandleWebhookRequest)
	http.HandleFunc(pathPrefix + "/fbm/webhook/", notFound) // TODO: Try to get rid?
	http.HandleFunc(pathPrefix + "/fbm/subscribe", h.Subscribe)
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
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		botCode := r.URL.Query().Get("bot")
		if botSettings, ok := h.botsBy.Code[botCode]; ok {
			var responseText string
			if q.Get("hub.verify_token") == botSettings.VerifyToken {
				responseText = q.Get("hub.challenge")
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				responseText = "Error, wrong validation token"
			}
			w.Write([]byte(responseText))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	case http.MethodPost:
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h FbmWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	var receivedMessage fbm_bot_api.ReceivedMessage
	log := h.BotHost.GetLogger(r)
	content := make([]byte, r.ContentLength)
	_, err = r.Body.Read(content)
	if err != nil {
		return
	}
	log.Infof("Request.Body: %v", string(content))
	err = json.Unmarshal(content, &receivedMessage)
	if err != nil {
		return
	}
	log.Infof("Unmarshaled JSON to a struct with %v entries: %v", len(receivedMessage.Entries), receivedMessage)
	entriesWithInputs = make([]bots.EntryInputs, len(receivedMessage.Entries))
	for i, entry := range receivedMessage.Entries {
		entryWithInputs := bots.EntryInputs{
			Entry:  entry,
			Inputs: make([]bots.WebhookInput, len(entry.Messaging)),
		}
		for j, messaging := range entry.Messaging {
			entryWithInputs.Inputs[j] = bots.WebhookInput(messaging)
		}
		entriesWithInputs[i] = entryWithInputs
	}
	botContext = bots.BotContext{
		BotHost: h.BotHost,
		//BotSettings: nil, // TODO: fill with actual
	}
	return
}

func (h FbmWebhookHandler) CreateWebhookContext(r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, translator bots.Translator) bots.WebhookContext {
	panic("Not implemented yet") //return NewTelegramWebhookContext(r, botContext, webhookInput)
}

func (h FbmWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	panic("Not implemented yet") //return NewTelegramWebhookResponder(w, r)
}
