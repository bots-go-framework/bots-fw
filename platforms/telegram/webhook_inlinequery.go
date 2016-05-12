package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookInlineQuery struct {
	updateID int
	inline tgbotapi.InlineQuery
}
var _ bots.WebhookInlineQuery = (*TelegramWebhookInlineQuery)(nil)

func NewTelegramWebhookInlineQuery(updateID int, message tgbotapi.Message) TelegramWebhookMessage {
	return TelegramWebhookMessage{updateID: updateID, message: message}
}

func (iq TelegramWebhookInlineQuery) GetID() interface{} {
	return iq.updateID
}

func (iq TelegramWebhookInlineQuery) Sequence() int {
	return iq.updateID
}

func (q TelegramWebhookInlineQuery) GetQuery() string {
	return q.inline.Query
}

func (iq TelegramWebhookInlineQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.inline.From}
}

func (iq TelegramWebhookInlineQuery) GetOffset() string {
	return iq.inline.Offset
}
