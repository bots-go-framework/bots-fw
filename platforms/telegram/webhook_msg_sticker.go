package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)


type TelegramWebhookStickerMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookStickerMessage = (*TelegramWebhookStickerMessage)(nil)

func (_ TelegramWebhookStickerMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputSticker
}

func NewTelegramWebhookStickerMessage(input telegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookStickerMessage {
	return TelegramWebhookStickerMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType: tgMessageType,
	}
}

//func (whm TelegramWebhookStickerMessage) IsEdited() bool {
//	return whm.TgMessageType == TelegramMessageTypeEdited || whm.TgMessageType == TelegramMessageTypeEditedChannelPost
//}
