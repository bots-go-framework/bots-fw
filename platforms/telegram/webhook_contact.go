package telegram

import (
	"github.com/strongo/bots-framework/core"
)

type tgWebhookContactMessage struct {
	tgWebhookMessage
}

func (tgWebhookContactMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputContact
}

var _ bots.WebhookContactMessage = (*tgWebhookContactMessage)(nil)

func newTgWebhookContact(input tgWebhookInput) tgWebhookContactMessage {
	return tgWebhookContactMessage{tgWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m tgWebhookContactMessage) FirstName() string {
	return m.update.Message.Contact.FirstName
}

func (m tgWebhookContactMessage) LastName() string {
	return m.update.Message.Contact.LastName
}

func (m tgWebhookContactMessage) PhoneNumber() string {
	return m.update.Message.Contact.PhoneNumber
}

func (m tgWebhookContactMessage) UserID() interface{} {
	return m.update.Message.Contact.UserID
}
