package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookChosenInlineResult struct {
	updateID           int
	chosenInlineResult *tgbotapi.ChosenInlineResult
}

var _ bots.WebhookChosenInlineResult = (*TelegramWebhookChosenInlineResult)(nil)

func NewTelegramWebhookChosenInlineResult(updateID int, chosenInlineResult *tgbotapi.ChosenInlineResult) TelegramWebhookChosenInlineResult {
	return TelegramWebhookChosenInlineResult{updateID: updateID, chosenInlineResult: chosenInlineResult}
}

func (q TelegramWebhookChosenInlineResult) GetID() interface{} {
	return q.updateID
}

func (q TelegramWebhookChosenInlineResult) GetResultID() string {
	return q.chosenInlineResult.ResultID
}

func (q TelegramWebhookChosenInlineResult) Sequence() int {
	return q.updateID
}

func (q TelegramWebhookChosenInlineResult) GetQuery() string {
	return q.chosenInlineResult.Query
}

func (q TelegramWebhookChosenInlineResult) GetInlineMessageID() string {
	return q.chosenInlineResult.InlineMessageID
}

func (iq TelegramWebhookChosenInlineResult) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.chosenInlineResult.From}
}
