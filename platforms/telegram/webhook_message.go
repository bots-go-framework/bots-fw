package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
)

type telegramWebhookMessage struct {
	telegramWebhookInput
	message *tgbotapi.Message // Can be either whi.update.Message or whi.update.CallbackQuery.Message
}

func (whm telegramWebhookMessage) IntID() int64 {
	return (int64)(whm.message.MessageID)
}

func newTelegramWebhookMessage(input telegramWebhookInput, message *tgbotapi.Message) telegramWebhookMessage {
	if message == nil {
		panic("message == nil")
	}
	return telegramWebhookMessage{telegramWebhookInput: input, message: message}
}
