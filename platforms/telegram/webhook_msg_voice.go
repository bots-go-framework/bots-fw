package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookVoiceMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookVoiceMessage = (*TelegramWebhookVoiceMessage)(nil)

func (_ TelegramWebhookVoiceMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputVoice
}

func NewTelegramWebhookVoiceMessage(input TelegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookVoiceMessage {
	return TelegramWebhookVoiceMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType: tgMessageType,
	}
}