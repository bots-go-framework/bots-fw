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
	if updateID == 0 {
		panic("updateID == 0")
	}
	if callbackQuery == nil {
		panic("callbackQuery == nil")
	}
	q := TelegramWebhookCallbackQuery{updateID: updateID, callbackQuery: callbackQuery}
	if callbackQuery.Message != nil && callbackQuery.Message.MessageID != 0 {
		q.message = NewTelegramWebhookMessage(updateID, callbackQuery.Message)
	}
	return q
}

func (iq TelegramWebhookCallbackQuery) GetID() interface{} {
	return iq.updateID
}

func (iq TelegramWebhookCallbackQuery) Chat() bots.WebhookChat {
	if iq.callbackQuery != nil && iq.callbackQuery.Message != nil {
		return TelegramWebhookChat{
			chat: iq.callbackQuery.Message.Chat,
		}
	}
	return nil
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

func EditMessageOnCallbackQuery(whcbq bots.WebhookCallbackQuery, parseMode, text string) *tgbotapi.EditMessageTextConfig {
	twhcbq := whcbq.(TelegramWebhookCallbackQuery)
	callbackQuery := twhcbq.callbackQuery

	emc := tgbotapi.EditMessageTextConfig{
		Text:      text,
		ParseMode: parseMode,
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: callbackQuery.InlineMessageID,
		},
	}
	if emc.InlineMessageID == "" {
		emc.ChatID = callbackQuery.Message.Chat.ID
		emc.MessageID = callbackQuery.Message.MessageID
	}
	return &emc
}
