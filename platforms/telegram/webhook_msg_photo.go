package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgWebhookPhotoMessage struct {
	tgWebhookMessage
	TgMessageType TgMessageType
}

var _ bots.WebhookPhotoMessage = (*tgWebhookPhotoMessage)(nil)

func (tgWebhookPhotoMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputPhoto
}

func newTgWebhookPhotoMessage(input tgWebhookInput, tgMessageType TgMessageType, tgMessage *tgbotapi.Message) tgWebhookPhotoMessage {
	return tgWebhookPhotoMessage{
		tgWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:    tgMessageType,
	}
}
