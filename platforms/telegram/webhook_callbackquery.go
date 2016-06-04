package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookCallbackQuery struct {
	updateID      int
	callbackQuery *tgbotapi.CallbackQuery
	message       bots.WebhookMessage
}

var _ bots.WebhookCallbackQuery = (*TelegramWebhookCallbackQuery)(nil)

func NewTelegramWebhookCallbackQuery(updateID int, callbackQuery *tgbotapi.CallbackQuery) TelegramWebhookCallbackQuery {
	q := TelegramWebhookCallbackQuery{updateID: updateID, callbackQuery: callbackQuery}
	if callbackQuery.Message.MessageID != 0 {
		q.message = NewTelegramWebhookMessage(updateID, callbackQuery.Message)
	}
	return q
}

func (iq TelegramWebhookCallbackQuery) GetID() interface{} {
	return iq.updateID
}

func (iq TelegramWebhookCallbackQuery) Sequence() int {
	return iq.updateID
}

func (q TelegramWebhookCallbackQuery) GetMessage() bots.WebhookMessage {
	return q.message
}

func (iq TelegramWebhookCallbackQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.callbackQuery.From}
}

func (iq TelegramWebhookCallbackQuery) GetData() string {
	return iq.callbackQuery.Data
}

func (iq TelegramWebhookCallbackQuery) GetInlineMessageID() string {
	return iq.callbackQuery.InlineMessageID
}
