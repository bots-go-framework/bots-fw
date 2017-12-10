package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookVoiceMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookVoiceMessage = (*TelegramWebhookVoiceMessage)(nil)

func (_ TelegramWebhookVoiceMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputVoice
}

func NewTelegramWebhookVoiceMessage(input telegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookVoiceMessage {
	return TelegramWebhookVoiceMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:          tgMessageType,
	}
}
