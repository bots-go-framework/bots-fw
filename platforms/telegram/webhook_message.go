package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
)

type telegramWebhookMessage struct {
	telegramWebhookInput
	message *tgbotapi.Message
}

func (whm telegramWebhookMessage) IntID() int64 {
	return (int64)(whm.update.Message.MessageID)
}

func newTelegramWebhookMessage(input telegramWebhookInput, message *tgbotapi.Message) telegramWebhookMessage {
	return telegramWebhookMessage{telegramWebhookInput: input, message: message}
}
