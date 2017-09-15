package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookAudioMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookAudioMessage = (*TelegramWebhookAudioMessage)(nil)

func (_ TelegramWebhookAudioMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputAudio
}

func NewTelegramWebhookAudioMessage(input TelegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookAudioMessage {
	return TelegramWebhookAudioMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType: tgMessageType,
	}
}