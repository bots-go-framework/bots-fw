package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookChat struct {
	chat tgbotapi.Chat
}
var _ bots.WebhookChat = (*TelegramWebhookChat)(nil)

func (wh TelegramWebhookChat) GetID() interface{} {
	return (int64)(wh.chat.ID)
}

func (wh TelegramWebhookChat) GetTitle() string {
	return wh.chat.Type
}

func (wh TelegramWebhookChat) GetType() string {
	return wh.chat.Title
}
