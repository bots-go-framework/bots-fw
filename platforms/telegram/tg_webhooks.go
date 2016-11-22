package telegram_bot

import (
	"encoding/json"
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
	"google.golang.org/appengine"
	"io/ioutil"
	"net/http"
	"github.com/pkg/errors"
)

func NewTelegramWebhookHandler(botsBy bots.BotSettingsProvider, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) TelegramWebhookHandler {
	if webhookDriver == nil {
		panic("webhookDriver == nil")
	}
	if botHost == nil {
		panic("botHost == nil")
	}
	if translatorProvider == nil {
		panic("translatorProvider == nil")
	}
	return TelegramWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:        TelegramPlatform{},
			BotHost:            botHost,
			WebhookDriver:      webhookDriver,
			TranslatorProvider: translatorProvider,
		},
	}
}

type TelegramWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsProvider
}
var _ bots.WebhookHandler = (*TelegramWebhookHandler)(nil)

func (h TelegramWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix+"/telegram/webhook", h.HandleWebhookRequest)
	http.HandleFunc(pathPrefix+"/telegram/webhook/", notFound)
	http.HandleFunc(pathPrefix+"/telegram/setwebhook", h.SetWebhook)
}

func (h TelegramWebhookHandler) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleWebhook(w, r, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h TelegramWebhookHandler) SetWebhook(w http.ResponseWriter, r *http.Request) {
	logger := h.Logger(r)
	client := h.GetHttpClient(r)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "Missing required parameter: code", http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	botSettings, ok := h.botsBy(c).Code[botCode]
	if !ok {
		http.Error(w, fmt.Sprintf("Bot not found by code: %v", botCode), http.StatusBadRequest)
		return
	}
	bot := tgbotapi.NewBotAPIWithClient(botSettings.Token, client)
	//bot.Debug = true

	webhookUrl := fmt.Sprintf("https://%v/bot/telegram/webhook?token=%v", r.Host, bot.Token)

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(webhookUrl)); err != nil {
		logger.Errorf(c, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Webhook set"))
	}
}

func (h TelegramWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext *bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	logger := h.BotHost.Logger(r)
	token := r.URL.Query().Get("token")
	c := appengine.NewContext(r) //TODO: Remove dependency on AppEngine, should be passed indside.
	botSettings, ok := h.botsBy(c).ApiToken[token]
	if !ok {
		errMess := fmt.Sprintf("Unknown token: [%v]", token)
		err = bots.AuthFailedError(errMess)
		return
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	if len(bytes) < 1024 * 3 {
		logger.Debugf(c, "Request body: %v", (string)(bytes))
	} else {
		logger.Debugf(c, "Request len(body): %v", len(bytes))
	}
	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			logger.Errorf(c, "json.UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
		} else if se, ok := err.(*json.SyntaxError); ok {
			logger.Errorf(c, "json.SyntaxError: Offset=%v", se.Offset)
		} else {
			logger.Errorf(c, "json.Error: %T: %v", err, err.Error())
		}
		return
	}
	botContext = bots.NewBotContext(h.BotHost, botSettings)
	input := NewTelegramWebhookInput(update)
	if input == nil {
		err = errors.New("Unexpected input")
		return
	}
	entriesWithInputs = []bots.EntryInputs{
			{
				Entry:  TelegramWebhookEntry{update: update},
				Inputs: []bots.WebhookInput{input},
			},
		}
	return
}

func (h TelegramWebhookHandler) CreateWebhookContext(appContext bots.BotAppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput, botCoreStores bots.BotCoreStores, gaMeasurement *measurement.BufferedSender) bots.WebhookContext {
	return NewTelegramWebhookContext(appContext, r, botContext, webhookInput, botCoreStores, gaMeasurement)
}

func (h TelegramWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if twhc, ok := whc.(*TelegramWebhookContext); ok {
		return NewTelegramWebhookResponder(w, twhc)
	} else {
		panic(fmt.Sprintf("Expected TelegramWebhookContext, got: %T", whc))
	}
}

func (h TelegramWebhookHandler) CreateBotCoreStores(appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(TelegramPlatformID, appContext, r)
}
