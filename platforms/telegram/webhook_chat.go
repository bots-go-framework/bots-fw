package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"strconv"
)

// TgWebhookChat is wrapper for Telegram chat
type TgWebhookChat struct {
	chat *tgbotapi.Chat
}

var _ bots.WebhookChat = (*TgWebhookChat)(nil)

// GetID returns telegram chat ID
func (wh TgWebhookChat) GetID() string {
	return strconv.FormatInt(wh.chat.ID, 10)
}

// GetType returns telegram chat type
func (wh TgWebhookChat) GetType() string {
	return wh.chat.Type
}

// IsGroupChat indicates type of chat (group or private)
func (wh TgWebhookChat) IsGroupChat() bool {
	return !wh.chat.IsPrivate()
}
