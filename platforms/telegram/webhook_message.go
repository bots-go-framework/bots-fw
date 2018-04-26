package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"strconv"
)

type tgWebhookMessage struct {
	tgWebhookInput
	message *tgbotapi.Message // Can be either whi.update.Message or whi.update.CallbackQuery.Message
}

func (whm tgWebhookMessage) IntID() int64 {
	return (int64)(whm.message.MessageID)
}

func newTelegramWebhookMessage(input tgWebhookInput, message *tgbotapi.Message) tgWebhookMessage {
	if message == nil {
		panic("message == nil")
	}
	return tgWebhookMessage{tgWebhookInput: input, message: message}
}

func (whm tgWebhookMessage) BotChatID() (string, error) {
	return strconv.FormatInt(whm.message.Chat.ID, 10), nil
}
