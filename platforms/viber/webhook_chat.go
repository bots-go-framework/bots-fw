package viber

import (
	"github.com/strongo/bots-framework/core"
)

type viberWebhookChat struct {
	viberUserID string
}

var _ bots.WebhookChat = (*viberWebhookChat)(nil)

func (wh viberWebhookChat) GetID() string {
	return wh.viberUserID
}

func (wh viberWebhookChat) GetType() string {
	return "private"
}

func (wh viberWebhookChat) IsGroupChat() bool {
	return false
}

func newViberWebhookChat(viberUserID string) viberWebhookChat {
	return viberWebhookChat{viberUserID: viberUserID}
}
