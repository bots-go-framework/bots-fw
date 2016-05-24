package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func NewTelegramWebhookHandler(botsBy bots.BotSettingsBy, webhookDriver bots.WebhookDriver, botHost bots.BotHost, translatorProvider bots.TranslatorProvider) TelegramWebhookHandler {
	return TelegramWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:   TelegramPlatform{},
			BotHost:       botHost,
			WebhookDriver: webhookDriver,
			TranslatorProvider: translatorProvider,
		},
	}
}

type TelegramWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsBy
}

func (h TelegramWebhookHandler) RegisterHandlers(pathPrefix string, notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(pathPrefix + "/telegram/webhook", h.HandleWebhookRequest)
	http.HandleFunc(pathPrefix + "/telegram/webhook/", notFound)
	http.HandleFunc(pathPrefix + "/telegram/setwebhook", h.SetWebhook)
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
	logger := h.GetLogger(r)
	client := h.GetHttpClient(r)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "Missing required parameter: code", http.StatusBadRequest)
		return
	}
	botSettings, ok := h.botsBy.Code[botCode]
	if !ok {
		http.Error(w, fmt.Sprintf("Bot not found by code: %v", botCode), http.StatusBadRequest)
		return
	}
	bot, err := tgbotapi.NewBotAPIWithClient(botSettings.Token, client)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create bot[%v]: %v", botCode, err), http.StatusInternalServerError)
		return
	}
	//bot.Debug = true

	webhookUrl := fmt.Sprintf("https://%v/bot/telegram/webhook?token=%v", r.Host, bot.Token)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(webhookUrl))
	if err != nil {
		logger.Errorf("%v", err)
	}
	w.Write([]byte("Webhook set"))
}

func (h TelegramWebhookHandler) GetBotContextAndInputs(r *http.Request) (botContext bots.BotContext, entriesWithInputs []bots.EntryInputs, err error) {
	logger := h.BotHost.GetLogger(r)
	token := r.URL.Query().Get("token")
	botSettings, ok := h.botsBy.ApiToken[token]
	if !ok {
		errMess := fmt.Sprintf("Unknown token: [%v]", token)
		err = bots.AuthFailedError(errMess)
		return
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	if len(bytes) < 1024 {
		logger.Debugf("Request body: %v", (string)(bytes))
	} else {
		logger.Debugf("Request len(body): %v", len(bytes))
	}
	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			logger.Errorf("json.UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
		} else if se, ok := err.(*json.SyntaxError); ok {
			logger.Errorf("json.SyntaxError: Offset=%v", se.Offset)
		} else {
			logger.Errorf("json.Error: %T: %v", err, err.Error())
		}
		return
	}
	return bots.BotContext{
		BotHost: h.BotHost,
		BotSettings: botSettings,
	},
	[]bots.EntryInputs{
		bots.EntryInputs{
			Entry: TelegramWebhookEntry{update: update},
			Inputs: []bots.WebhookInput{NewTelegramWebhookInput(update)},
		},
	},
	nil
}

func (h TelegramWebhookHandler) CreateWebhookContext(appContext bots.AppContext, r *http.Request, botContext bots.BotContext, webhookInput bots.WebhookInput,  botCoreStores bots.BotCoreStores) bots.WebhookContext {
	return NewTelegramWebhookContext(appContext, r, botContext, webhookInput, botCoreStores)
}

func (h TelegramWebhookHandler) GetResponder(w http.ResponseWriter, whc bots.WebhookContext) bots.WebhookResponder {
	if twhc, ok := whc.(*TelegramWebhookContext); ok {
		return NewTelegramWebhookResponder(w, twhc)
	} else {
		panic(fmt.Sprintf("Expected TelegramWebhookContext, got: %T", whc))
	}
}

func (h TelegramWebhookHandler) CreateBotCoreStores (appContext bots.AppContext, r *http.Request) bots.BotCoreStores {
	return h.BotHost.GetBotCoreStores(TelegramPlatformID, appContext, r)
}

