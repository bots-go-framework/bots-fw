package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgWebhookAudioMessage struct {
	tgWebhookMessage
	TgMessageType TgMessageType
}

var _ bots.WebhookAudioMessage = (*tgWebhookAudioMessage)(nil)

func (tgWebhookAudioMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputAudio
}

func newTgWebhookAudioMessage(input tgWebhookInput, tgMessageType TgMessageType, tgMessage *tgbotapi.Message) tgWebhookAudioMessage {
	return tgWebhookAudioMessage{
		tgWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:    tgMessageType,
	}
}
