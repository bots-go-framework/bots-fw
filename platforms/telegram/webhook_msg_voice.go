package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgWebhookVoiceMessage struct {
	tgWebhookMessage
	TgMessageType TgMessageType
}

var _ bots.WebhookVoiceMessage = (*tgWebhookVoiceMessage)(nil)

func (tgWebhookVoiceMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputVoice
}

func newTgWebhookVoiceMessage(input tgWebhookInput, tgMessageType TgMessageType, tgMessage *tgbotapi.Message) tgWebhookVoiceMessage {
	return tgWebhookVoiceMessage{
		tgWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:    tgMessageType,
	}
}
