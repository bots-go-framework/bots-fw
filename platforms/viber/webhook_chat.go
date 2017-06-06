package viber_bot

import (
	"github.com/strongo/bots-framework/core"
)

type ViberWebhookChat struct {
	viberUserID string
}

var _ bots.WebhookChat = (*ViberWebhookChat)(nil)

func (wh ViberWebhookChat) GetID() string {
	return wh.viberUserID
}

func (wh ViberWebhookChat) GetType() string {
	return "private"
}

func (wh ViberWebhookChat) IsGroupChat() bool {
	return false
}

func NewViberWebhookChat(viberUserID string) ViberWebhookChat {
	return ViberWebhookChat{viberUserID: viberUserID}
}