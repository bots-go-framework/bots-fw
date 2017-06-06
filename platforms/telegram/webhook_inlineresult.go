package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookChosenInlineResult struct {
	TelegramWebhookInput
}

var _ bots.WebhookChosenInlineResult = (*TelegramWebhookChosenInlineResult)(nil)

func (_ TelegramWebhookChosenInlineResult) InputType() bots.WebhookInputType {
	return bots.WebhookInputChosenInlineResult
}

func NewTelegramWebhookChosenInlineResult(input TelegramWebhookInput) TelegramWebhookChosenInlineResult {
	return TelegramWebhookChosenInlineResult{TelegramWebhookInput: input}
}


func (q TelegramWebhookChosenInlineResult) GetResultID() string {
	return q.update.ChosenInlineResult.ResultID
}

func (q TelegramWebhookChosenInlineResult) GetQuery() string {
	return q.update.ChosenInlineResult.Query
}

func (q TelegramWebhookChosenInlineResult) GetInlineMessageID() string {
	return q.update.ChosenInlineResult.InlineMessageID
}

func (iq TelegramWebhookChosenInlineResult) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.update.ChosenInlineResult.From}
}
