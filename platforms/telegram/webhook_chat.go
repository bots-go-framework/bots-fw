package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"strconv"
)

type TelegramWebhookChat struct {
	chat *tgbotapi.Chat
}

var _ bots.WebhookChat = (*TelegramWebhookChat)(nil)

func (wh TelegramWebhookChat) GetID() string {
	return strconv.FormatInt(wh.chat.ID, 10)
}

func (wh TelegramWebhookChat) GetFullName() string {
	return wh.chat.Type
}

func (wh TelegramWebhookChat) GetType() string {
	return wh.chat.Title
}
