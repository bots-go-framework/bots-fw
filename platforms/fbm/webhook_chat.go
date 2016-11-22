package fbm_bot

import "github.com/strongo/bots-framework/core"

type FbmWebhookChat struct {
	ID string
}

var _ bots.WebhookChat = (*FbmWebhookChat)(nil)

func (wh FbmWebhookChat) GetID() string {
	return wh.ID
}

func (wh FbmWebhookChat) GetFullName() string {
	return "not implemented"
}

func (wh FbmWebhookChat) GetType() string {
	return "not implemented"
}

func NewFbmWebhookChat(id string) FbmWebhookChat {
	return FbmWebhookChat{
		ID: id,
	}
}