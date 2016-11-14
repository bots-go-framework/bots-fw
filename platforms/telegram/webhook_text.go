package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookTextMessage struct {
	telegramWebhookMessage
}

var _ bots.WebhookTextMessage = (*TelegramWebhookTextMessage)(nil)

func (_ TelegramWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func NewTelegramWebhookTextMessage(input telegramWebhookInput) TelegramWebhookTextMessage {
	return TelegramWebhookTextMessage{telegramWebhookMessage: telegramWebhookMessage{telegramWebhookInput: input}}
}

func (whm TelegramWebhookTextMessage) Text() string {
	return whm.update.Message.Text
}

