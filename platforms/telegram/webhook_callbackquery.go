package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"strconv"
)

type TelegramWebhookCallbackQuery struct {
	telegramWebhookInput
	//callbackQuery *tgbotapi.CallbackQuery
	//message       bots.WebhookMessage
}

var (
	_ bots.WebhookCallbackQuery = (*TelegramWebhookCallbackQuery)(nil)
	_ TelegramWebhookInput      = (*TelegramWebhookCallbackQuery)(nil)
	_ bots.WebhookInput         = (*TelegramWebhookCallbackQuery)(nil)
)

func (twhcbq TelegramWebhookCallbackQuery) InputType() bots.WebhookInputType {
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

func (twhcbq TelegramWebhookCallbackQuery) GetID() interface{} {
	return twhcbq.update.UpdateID
}

func (twhcbq TelegramWebhookCallbackQuery) Sequence() int {
	return twhcbq.update.UpdateID
}

func (twhcbq TelegramWebhookCallbackQuery) GetMessage() bots.WebhookMessage {
	return newTelegramWebhookMessage(twhcbq.telegramWebhookInput, twhcbq.update.CallbackQuery.Message)
}

func (twhcbq TelegramWebhookCallbackQuery) TelegramCallbackMessage() *tgbotapi.Message {
	return twhcbq.update.CallbackQuery.Message
}

func (twhcbq TelegramWebhookCallbackQuery) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: twhcbq.update.CallbackQuery.From}
}

func (twhcbq TelegramWebhookCallbackQuery) GetData() string {
	return twhcbq.update.CallbackQuery.Data
}

func (twhcbq TelegramWebhookCallbackQuery) GetInlineMessageID() string {
	return twhcbq.update.CallbackQuery.InlineMessageID
}

func (twhcbq TelegramWebhookCallbackQuery) BotChatID() (string, error) {
	if cbq := twhcbq.update.CallbackQuery; cbq.Message != nil && cbq.Message.Chat != nil {
		return strconv.FormatInt(cbq.Message.Chat.ID, 10), nil
	}
	return "", nil
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
