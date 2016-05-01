package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"fmt"
	"net/http"
)

func NewTelegramWebhookHandler(botsBy bots.BotSettingsBy, webhookDriver bots.WebhookDriver, botHost bots.BotHost) TelegramWebhookHandler {
	return TelegramWebhookHandler{
		botsBy: botsBy,
		BaseHandler: bots.BaseHandler{
			BotPlatform: TelegramPlatform{},
			BotHost: botHost,
			WebhookDriver: webhookDriver,
		},
	}
}

type TelegramWebhookHandler struct {
	bots.BaseHandler
	botsBy bots.BotSettingsBy
}

func (h TelegramWebhookHandler) RegisterHandlers(notFound func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc("/bot/telegram/webhook", h.HandleRequest)
	http.HandleFunc("/bot/telegram/webhook/", notFound)
	http.HandleFunc("/bot/telegram/setwebhook", h.SetWebhook)
}

func (h TelegramWebhookHandler) HandleRequest (w http.ResponseWriter, r *http.Request) {
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

func (h TelegramWebhookHandler) GetEntryInputs(r *http.Request) (entriesWithInputs []bots.EntryInputs, err error) {
	return
}

