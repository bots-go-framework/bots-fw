package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgWebhookStickerMessage struct {
	tgWebhookMessage
	TgMessageType TgMessageType
}

var _ bots.WebhookStickerMessage = (*tgWebhookStickerMessage)(nil)

func (tgWebhookStickerMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputSticker
}

func newTgWebhookStickerMessage(input tgWebhookInput, tgMessageType TgMessageType, tgMessage *tgbotapi.Message) tgWebhookStickerMessage {
	return tgWebhookStickerMessage{
		tgWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType:    tgMessageType,
	}
}

//func (whm tgWebhookStickerMessage) IsEdited() bool {
//	return whm.TgMessageType == TgMessageTypeEdited || whm.TgMessageType == TgMessageTypeEditedChannelPost
//}
