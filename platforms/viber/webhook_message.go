package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)


type ViberWebhookMessage struct {
	message  *viberinterface.CallbackMessage
}

var _ bots.WebhookMessage = (*ViberWebhookMessage)(nil)

func NewViberWebhookMessage(message *viberinterface.CallbackMessage) ViberWebhookMessage {
	return ViberWebhookMessage{message: message}
}

func (whm ViberWebhookMessage) IntID() int64 {
	return 0
}

func (whm ViberWebhookMessage) Contact() bots.WebhookContact {
	//if whm.message.Contact != nil {
	//	return NewViberWebhookContact(whm.message.Contact)
	//}
	return nil
}

func (whm ViberWebhookMessage) StringID() string {
	return ""
}

func (whm ViberWebhookMessage) Sequence() int {
	return 0
}

func (whm ViberWebhookMessage) Text() string {
	return whm.message.Text
}
