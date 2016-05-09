package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"time"
)

type TelegramWebhookInput struct {
	inputType bots.WebhookInputType
	update tgbotapi.Update
}
var _ bots.WebhookInput = (*TelegramWebhookInput)(nil)

func NewTelegramWebhookInput(update tgbotapi.Update) TelegramWebhookInput {
	result := TelegramWebhookInput{update: update}
	switch {
	case update.Message.MessageID > 0: result.inputType = bots.WebhookInputMessage
	case update.InlineQuery.ID != "": result.inputType = bots.WebhookInputInlineQuery
	case update.ChosenInlineResult.ResultID != "": result.inputType = bots.WebhookInputChoosenInlineResult
	}
	return result
}

func (whi TelegramWebhookInput) GetSender() bots.WebhookSender{
	return TelegramSender{tgUser: whi.update.Message.From}
}

func (whi TelegramWebhookInput) GetRecipient() bots.WebhookRecipient {
	panic("Not implemented")
}

func (whi TelegramWebhookInput) GetTime() time.Time {
	return whi.update.Message.Time()
}

func (whi TelegramWebhookInput) InputType() bots.WebhookInputType {
	return whi.inputType
}

func (whi TelegramWebhookInput) InputMessage() bots.WebhookMessage {
	update := whi.update
	return NewTelegramWebhookMessage(update.UpdateID, update.Message)
}

func (whi TelegramWebhookInput) InputPostback() bots.WebhookPostback {
	panic("Not implemented")
}

func (whi TelegramWebhookInput) InputDelivery() bots.WebhookDelivery {
	panic("Not implemented")
}
