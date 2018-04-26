package viber

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

// viberWebhookTextMessage is Viber text message
type viberWebhookTextMessage struct {
	viberWebhookMessage
}

var _ bots.WebhookTextMessage = (*viberWebhookTextMessage)(nil)

func (viberWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func newViberWebhookTextMessage(m viberinterface.CallbackOnMessage) viberWebhookTextMessage {
	return viberWebhookTextMessage{viberWebhookMessage: newViberWebhookMessage(m)}
}

func (whm viberWebhookTextMessage) Text() string {
	return whm.m.Message.Text
}

func (whm viberWebhookTextMessage) IsEdited() bool {
	return false
}
