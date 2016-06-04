package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookInlineQuery struct {
	updateID    int
	inlineQuery *tgbotapi.InlineQuery
}

var _ bots.WebhookInlineQuery = (*TelegramWebhookInlineQuery)(nil)

func NewTelegramWebhookInlineQuery(updateID int, inlineQuery *tgbotapi.InlineQuery) TelegramWebhookInlineQuery {
	return TelegramWebhookInlineQuery{updateID: updateID, inlineQuery: inlineQuery}
}

func (q TelegramWebhookInlineQuery) GetID() interface{} {
	return q.updateID
}

func (q TelegramWebhookInlineQuery) GetInlineQueryID() string {
	return q.inlineQuery.ID
}

func (q TelegramWebhookInlineQuery) Sequence() int {
	return q.updateID
}

func (q TelegramWebhookInlineQuery) GetQuery() string {
	return q.inlineQuery.Query
}

func (iq TelegramWebhookInlineQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.inlineQuery.From}
}

func (iq TelegramWebhookInlineQuery) GetOffset() string {
	return iq.inlineQuery.Offset
}
