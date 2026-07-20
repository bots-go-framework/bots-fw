package botsfw

import (
	"github.com/strongo/analytics"
	"strings"
)

type WebhookAnalytics interface {
	Enqueue(message analytics.Message)
}

var _ WebhookAnalytics = (*webhookAnalytics)(nil)

type webhookAnalytics struct {
	whcb *WebhookContextBase
}

func (wha webhookAnalytics) UserContext() analytics.UserContext {
	var userLanguage string
	if chatData := wha.whcb.ChatData(); chatData != nil {
		userLanguage = strings.ToLower(chatData.GetPreferredLanguage())
	}
	return analytics.NewUserContext(wha.whcb.AppUserID()).SetUserLanguage(userLanguage)
}

func (wha webhookAnalytics) Enqueue(message analytics.Message) {
	ctx := wha.whcb.Context()
	wha.UserContext().QueueMessage(ctx, message)
}
