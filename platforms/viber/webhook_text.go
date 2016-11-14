package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-viber/viberinterface"
)

type ViberWebhookTextMessage struct {
	viberWebhookMessage
}

var _ bots.WebhookTextMessage = (*ViberWebhookTextMessage)(nil)

func (_ ViberWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func NewViberWebhookTextMessage(m viberinterface.CallbackOnMessage) ViberWebhookTextMessage {
	return ViberWebhookTextMessage{viberWebhookMessage: newViberWebhookMessage(m)}
}

func (whm ViberWebhookTextMessage) Text() string {
	return whm.m.Message.Text
}

