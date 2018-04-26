package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgWebhookTextMessage struct {
	tgWebhookMessage
	TgMessageType TgMessageType
}

var _ bots.WebhookTextMessage = (*tgWebhookTextMessage)(nil)

func (tgWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func newTgWebhookTextMessage(input tgWebhookInput, tgMessageType TgMessageType, tgMessage *tgbotapi.Message) tgWebhookTextMessage {
	return tgWebhookTextMessage{
		tgWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:    tgMessageType,
	}
}

func (whm tgWebhookTextMessage) Text() string {
	return whm.message.Text
}

func (whm tgWebhookTextMessage) IsEdited() bool {
	return whm.TgMessageType == TgMessageTypeEdited || whm.TgMessageType == TgMessageTypeEditedChannelPost
}
