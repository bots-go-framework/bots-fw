package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookContact struct {
	contact *tgbotapi.Contact
}

var _ bots.WebhookContact = (*TelegramWebhookContact)(nil)

func NewTelegramWebhookContact(contact *tgbotapi.Contact) TelegramWebhookContact {
	return TelegramWebhookContact{contact: contact}
}

func (m TelegramWebhookContact) FirstName() string {
	return m.contact.FirstName
}

func (m TelegramWebhookContact) LastName() string {
	return m.contact.LastName
}

func (m TelegramWebhookContact) PhoneNumber() string {
	return m.contact.PhoneNumber
}

func (m TelegramWebhookContact) UserID() interface{} {
	return m.contact.UserID
}
