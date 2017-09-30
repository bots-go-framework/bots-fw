package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"golang.org/x/net/context"
	"strconv"
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

func (whm telegramWebhookMessage) BotChatID(c context.Context) (chatID string, err error) {
	return strconv.FormatInt(whm.message.Chat.ID, 10), nil
}