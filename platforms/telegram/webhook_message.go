package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookMessage struct {
	updateID int
	message tgbotapi.Message
}

func NewTelegramWebhookMessage(updateID int, message tgbotapi.Message) TelegramWebhookMessage {
	return TelegramWebhookMessage{updateID: updateID, message: message}
}
var _ bots.WebhookMessage = (*TelegramWebhookMessage)(nil)

func (whm TelegramWebhookMessage) IntID() int64 {
	return (int64)(whm.message.MessageID)
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
