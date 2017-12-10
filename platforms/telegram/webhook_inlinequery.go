package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookInlineQuery struct {
	telegramWebhookInput
}

func (_ TelegramWebhookInlineQuery) InputType() bots.WebhookInputType {
	return bots.WebhookInputInlineQuery
}

var _ bots.WebhookInlineQuery = (*TelegramWebhookInlineQuery)(nil)

func NewTelegramWebhookInlineQuery(input telegramWebhookInput) TelegramWebhookInlineQuery {
	return TelegramWebhookInlineQuery{telegramWebhookInput: input}
}

func (q TelegramWebhookInlineQuery) GetInlineQueryID() string {
	return q.update.InlineQuery.ID
}

func (q TelegramWebhookInlineQuery) GetQuery() string {
	return q.update.InlineQuery.Query
}

func (iq TelegramWebhookInlineQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.update.InlineQuery.From}
}

func (iq TelegramWebhookInlineQuery) GetOffset() string {
	return iq.update.InlineQuery.Offset
}

func (iq TelegramWebhookInlineQuery) BotChatID() (string, error) {
	return "", nil
}
