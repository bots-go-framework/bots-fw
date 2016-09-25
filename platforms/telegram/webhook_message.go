package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookMessage struct {
	updateID int
	message  *tgbotapi.Message
}

var _ bots.WebhookMessage = (*TelegramWebhookMessage)(nil)

func NewTelegramWebhookMessage(updateID int, message *tgbotapi.Message) TelegramWebhookMessage {
	if updateID == 0 {
		panic("updateID == 0")
	}
	if message == nil {
		panic("message == nil")
	}
	return TelegramWebhookMessage{updateID: updateID, message: message}
}

func (whm TelegramWebhookMessage) IntID() int64 {
	return (int64)(whm.message.MessageID)
}

func (whm TelegramWebhookMessage) Contact() bots.WebhookContact {
	if whm.message.Contact != nil {
		return NewTelegramWebhookContact(whm.message.Contact)
	}
	return nil
}

func (whm TelegramWebhookMessage) StringID() string {
	return ""
}

func (whm TelegramWebhookMessage) Sequence() int {
	return whm.updateID
}

func (whm TelegramWebhookMessage) Text() string {
	return whm.message.Text
}
