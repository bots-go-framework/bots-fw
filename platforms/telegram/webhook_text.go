package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookTextMessage struct {
	telegramWebhookMessage
	isEdited bool
}

var _ bots.WebhookTextMessage = (*TelegramWebhookTextMessage)(nil)

func (_ TelegramWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func NewTelegramWebhookTextMessage(input TelegramWebhookInput) TelegramWebhookTextMessage {
	m := input.update.Message
	var isEdited bool
	if m == nil {
		m = input.update.EditedMessage
		if m == nil {
			panic("Telegram update does not have Message or EditedMessage")
		}
		isEdited = true
	}
	return TelegramWebhookTextMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, m),
		isEdited: isEdited,
	}
}

func (whm TelegramWebhookTextMessage) Text() string {
	return whm.message.Text
}

func (whm TelegramWebhookTextMessage) IsEdited() bool {
	return whm.isEdited
}
