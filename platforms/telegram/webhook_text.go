package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
)

type TelegramMessageType string
const (
	TelegramMessageTypeRegular = "message"
	TelegramMessageTypeEdited = "edited_message"
	TelegramMessageTypeChannelPost = "channel_post"
	TelegramMessageTypeEditedChannelPost = "edited_channel_post"
)

type TelegramWebhookTextMessage struct {
	telegramWebhookMessage
	TgMessageType TelegramMessageType
}

var _ bots.WebhookTextMessage = (*TelegramWebhookTextMessage)(nil)

func (_ TelegramWebhookTextMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputText
}

func NewTelegramWebhookTextMessage(input TelegramWebhookInput, tgMessageType TelegramMessageType, tgMessage *tgbotapi.Message) TelegramWebhookTextMessage {
	return TelegramWebhookTextMessage{
		telegramWebhookMessage: newTelegramWebhookMessage(input, tgMessage),
		TgMessageType: tgMessageType,
	}
}

func (whm TelegramWebhookTextMessage) Text() string {
	return whm.message.Text
}

func (whm TelegramWebhookTextMessage) IsEdited() bool {
	return whm.TgMessageType == TelegramMessageTypeEdited || whm.TgMessageType == TelegramMessageTypeEditedChannelPost
}
