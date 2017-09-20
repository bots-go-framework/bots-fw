package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookCallbackQuery struct {
	telegramWebhookInput
	//callbackQuery *tgbotapi.CallbackQuery
	//message       bots.WebhookMessage
}

var (
	_ bots.WebhookCallbackQuery = (*TelegramWebhookCallbackQuery)(nil)
	_ TelegramWebhookInput = (*TelegramWebhookCallbackQuery)(nil)
)

func (_ TelegramWebhookCallbackQuery) InputType() bots.WebhookInputType {
	return bots.WebhookInputCallbackQuery
}

func NewTelegramWebhookCallbackQuery(input telegramWebhookInput) TelegramWebhookCallbackQuery {
	callbackQuery := input.update.CallbackQuery
	if callbackQuery == nil {
		panic("update.CallbackQuery == nil")
	}
	q := TelegramWebhookCallbackQuery{
		telegramWebhookInput: input,
	}
	return q
}

func (iq TelegramWebhookCallbackQuery) GetID() interface{} {
	return iq.update.UpdateID
}

func (iq TelegramWebhookCallbackQuery) Sequence() int {
	return iq.update.UpdateID
}

func (q TelegramWebhookCallbackQuery) GetMessage() bots.WebhookMessage {
	return newTelegramWebhookMessage(q.telegramWebhookInput, q.update.CallbackQuery.Message)
}

func (q TelegramWebhookCallbackQuery) TelegramCallbackMessage() *tgbotapi.Message {
	return q.update.CallbackQuery.Message
}

func (iq TelegramWebhookCallbackQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.update.CallbackQuery.From}
}

func (iq TelegramWebhookCallbackQuery) GetData() string {
	return iq.update.CallbackQuery.Data
}

func (iq TelegramWebhookCallbackQuery) GetInlineMessageID() string {
	return iq.update.CallbackQuery.InlineMessageID
}

func EditMessageOnCallbackQuery(whcbq bots.WebhookCallbackQuery, parseMode, text string) *tgbotapi.EditMessageTextConfig {
	twhcbq := whcbq.(TelegramWebhookCallbackQuery)
	callbackQuery := twhcbq.update.CallbackQuery

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
