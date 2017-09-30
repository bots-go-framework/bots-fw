package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

type TelegramWebhookChosenInlineResult struct {
	telegramWebhookInput
}

var _ bots.WebhookChosenInlineResult = (*TelegramWebhookChosenInlineResult)(nil)

func (_ TelegramWebhookChosenInlineResult) InputType() bots.WebhookInputType {
	return bots.WebhookInputChosenInlineResult
}

func NewTelegramWebhookChosenInlineResult(input telegramWebhookInput) TelegramWebhookChosenInlineResult {
	return TelegramWebhookChosenInlineResult{telegramWebhookInput: input}
}


func (q TelegramWebhookChosenInlineResult) GetResultID() string {
	return q.update.ChosenInlineResult.ResultID
}

func (q TelegramWebhookChosenInlineResult) GetQuery() string {
	return q.update.ChosenInlineResult.Query
}

func (q TelegramWebhookChosenInlineResult) GetInlineMessageID() string {
	if q.update.ChosenInlineResult != nil {
		return q.update.ChosenInlineResult.InlineMessageID
	}
	return ""
}

func (iq TelegramWebhookChosenInlineResult) GetFrom() bots.WebhookSender {
	return TelegramSender{tgUser: iq.update.ChosenInlineResult.From}
}

func (q TelegramWebhookChosenInlineResult) BotChatID(c context.Context) (chatID string, err error) {
	return "", nil
}