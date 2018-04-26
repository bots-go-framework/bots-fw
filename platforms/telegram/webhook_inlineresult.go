package telegram

import (
	"github.com/strongo/bots-framework/core"
)

type tgWebhookChosenInlineResult struct {
	tgWebhookInput
}

var _ bots.WebhookChosenInlineResult = (*tgWebhookChosenInlineResult)(nil)

func (tgWebhookChosenInlineResult) InputType() bots.WebhookInputType {
	return bots.WebhookInputChosenInlineResult
}

func newTelegramWebhookChosenInlineResult(input tgWebhookInput) tgWebhookChosenInlineResult {
	return tgWebhookChosenInlineResult{tgWebhookInput: input}
}

func (q tgWebhookChosenInlineResult) GetResultID() string {
	return q.update.ChosenInlineResult.ResultID
}

func (q tgWebhookChosenInlineResult) GetQuery() string {
	return q.update.ChosenInlineResult.Query
}

func (q tgWebhookChosenInlineResult) GetInlineMessageID() string {
	if q.update.ChosenInlineResult != nil {
		return q.update.ChosenInlineResult.InlineMessageID
	}
	return ""
}

func (q tgWebhookChosenInlineResult) GetFrom() bots.WebhookSender {
	return tgSender{tgUser: q.update.ChosenInlineResult.From}
}

func (q tgWebhookChosenInlineResult) BotChatID() (string, error) {
	return "", nil
}
