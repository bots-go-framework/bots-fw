package viber

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/log"
	"net/http"
	"net/url"
)

func (h viberWebhookHandler) SetWebhook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := h.Context(r)
	client := h.GetHTTPClient(c)
	botCode := r.URL.Query().Get("code")
	if botCode == "" {
		http.Error(w, "Missing required parameter: code", http.StatusBadRequest)
		return
	}
	botSettings, ok := h.botsBy(c).ByCode[botCode]
	if !ok {
		m := fmt.Sprintf("Bot not found by code: %v", botCode)
		http.Error(w, m, http.StatusBadRequest)
		log.Errorf(c, fmt.Sprintf("%v. All bots: %v", m, h.botsBy(c).ByCode))
		return
	}
	bot := viberbotapi.NewViberBotAPIWithHTTPClient(botSettings.Token, client)
	//bot.Debug = true

	webhookURL := fmt.Sprintf("https://%v/bot/viber/callback/%v", r.Host, url.QueryEscape(botSettings.Code))

	//eventTypes := []string {"delivered", "seen", "failed", "subscribed",  "unsubscribed", "conversation_started"}
	eventTypes := []string{"failed", "subscribed", "unsubscribed", "conversation_started"}

	if _, err := bot.SetWebhook(webhookURL, eventTypes); err != nil {
		log.Errorf(c, "%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Webhook set"))
	}
}
