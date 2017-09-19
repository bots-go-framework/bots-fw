package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookPhotoMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookPhotoMessage = (*TelegramWebhookPhotoMessage)(nil)

func (_ TelegramWebhookPhotoMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputPhoto
}

func NewTelegramWebhookPhotoMessage(input telegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookPhotoMessage {
	return TelegramWebhookPhotoMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType: tgMessageType,
	}
}