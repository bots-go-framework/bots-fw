package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookContactMessage struct {
	telegramWebhookMessage
}

func (_ TelegramWebhookContactMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputContact
}

var _ bots.WebhookContactMessage = (*TelegramWebhookContactMessage)(nil)

func NewTelegramWebhookContact(input telegramWebhookInput) TelegramWebhookContactMessage {
	return TelegramWebhookContactMessage{telegramWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m TelegramWebhookContactMessage) FirstName() string {
	return m.update.Message.Contact.FirstName
}

func (m TelegramWebhookContactMessage) LastName() string {
	return m.update.Message.Contact.LastName
}

func (m TelegramWebhookContactMessage) PhoneNumber() string {
	return m.update.Message.Contact.PhoneNumber
}

func (m TelegramWebhookContactMessage) UserID() interface{} {
	return m.update.Message.Contact.UserID
}
