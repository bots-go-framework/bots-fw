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

func (wha webhookAnalytics) UserContext() *analytics.UserContext {
	return &analytics.UserContext{
		UserID:       wha.whcb.AppUserID(),
		UserLanguage: strings.ToLower(wha.whcb.botChat.Data.GetPreferredLanguage()),
	}
}

func (wha webhookAnalytics) Enqueue(message analytics.Message) {
	ctx := wha.whcb.Context()
	wha.UserContext().QueueMessage(ctx, message)
}
