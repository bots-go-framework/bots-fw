package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func NewTelegramWebhookHandler(botsBy bots.BotSettingsBy, webhookDriver bots.WebhookDriver, botHost bots.BotHost) TelegramWebhookHandler {
	return TelegramWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform:   TelegramPlatform{},
			BotHost:       botHost,
			WebhookDriver: webhookDriver,
		},
	}
}

type TelegramWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsBy
}

func (h TelegramWebhookHandler) RegisterHandlers(notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc("/bot/telegram/webhook", h.HandleWebhookRequest)
	http.HandleFunc("/bot/telegram/webhook/", notFound)
	http.HandleFunc("/bot/telegram/setwebhook", h.SetWebhook)
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
	log := h.GetLogger(r)
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
		log.Errorf("%v", err)
	}
	w.Write([]byte("Webhook set"))
}

func (h TelegramWebhookHandler) GetBotContext(r *http.Request) (botContext bots.BotContext, err error) {
	token := r.URL.Query().Get("token")
	botSettings, ok := h.botsBy.ApiToken[token]
	if !ok {
		errMess := fmt.Sprintf("Unknown token: [%v]", token)
		return botContext, bots.AuthFailedError(errMess)
	}
	bytes, _ := ioutil.ReadAll(r.Body)
	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		return botContext, err
	}
	return bots.BotContext{
		BotSettings: botSettings,
		EntriesWithInputs: []bots.EntryInputs{
			bots.EntryInputs{
				Entry: TelegramWebhookEntry{update: update},
				Inputs: []bots.WebhookInput{NewTelegramWebhookInput(update)},
			},
		},
	}, nil
}

