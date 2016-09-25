package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"time"
)

type TelegramWebhookInput struct {
	inputType bots.WebhookInputType
	update    tgbotapi.Update
}

var _ bots.WebhookInput = (*TelegramWebhookInput)(nil)

func NewTelegramWebhookInput(update tgbotapi.Update) TelegramWebhookInput {
	result := TelegramWebhookInput{update: update}
	switch {
	case update.Message != nil:
		result.inputType = bots.WebhookInputMessage
	case update.InlineQuery != nil:
		result.inputType = bots.WebhookInputInlineQuery
	case update.CallbackQuery != nil:
		result.inputType = bots.WebhookInputCallbackQuery
	case update.ChosenInlineResult != nil:
		result.inputType = bots.WebhookInputChosenInlineResult
	}
	return result
}

func (whi TelegramWebhookInput) Chat() bots.WebhookChat {
	return TelegramWebhookChat{
		chat: whi.update.Message.Chat,
	}
}


func (whi TelegramWebhookInput) GetSender() bots.WebhookSender {
	switch whi.InputType() {
	case bots.WebhookInputMessage:
		return TelegramSender{tgUser: whi.update.Message.From}
	case bots.WebhookInputChosenInlineResult:
		return TelegramSender{tgUser: whi.update.ChosenInlineResult.From}
	case bots.WebhookInputInlineQuery:
		return TelegramSender{tgUser: whi.update.InlineQuery.From}
	case bots.WebhookInputCallbackQuery:
		return TelegramSender{tgUser: whi.update.CallbackQuery.From}
	}
	return nil
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
	if update.Message == nil {
		return nil // panic(fmt.Sprintf("Telegram update(id=%v).Message == nil", update.UpdateID))
	}
	return NewTelegramWebhookMessage(update.UpdateID, update.Message)
}

func (whi TelegramWebhookInput) InputInlineQuery() bots.WebhookInlineQuery {
	update := whi.update
	if update.InlineQuery == nil {
		panic(fmt.Sprintf("Telegram update(id=%v).InlineQuery == nil", update.UpdateID))
	}
	return NewTelegramWebhookInlineQuery(update.UpdateID, update.InlineQuery)
}

func (whi TelegramWebhookInput) InputChosenInlineResult() bots.WebhookChosenInlineResult {
	update := whi.update
	if update.ChosenInlineResult == nil {
		panic(fmt.Sprintf("Telegram update(id=%v).ChosenInlineResult == nil", update.UpdateID))
	}
	return NewTelegramWebhookChosenInlineResult(update.UpdateID, update.ChosenInlineResult)
}

func (whi TelegramWebhookInput) InputCallbackQuery() bots.WebhookCallbackQuery {
	update := whi.update
	if update.CallbackQuery == nil {
		panic(fmt.Sprintf("Telegram update(id=%v).CallbackQuery == nil", update.UpdateID))
	} else {
		_ = fmt.Sprintf("%v", update.UpdateID)
		_ = fmt.Sprintf("%v", update.CallbackQuery)
		_ = fmt.Sprintf("%v", update.CallbackQuery.ID)
		return NewTelegramWebhookCallbackQuery(update.UpdateID, update.CallbackQuery)
	}
}

func (whi TelegramWebhookInput) InputPostback() bots.WebhookPostback {
	panic("Not implemented")
}

func (whi TelegramWebhookInput) InputDelivery() bots.WebhookDelivery {
	panic("Not implemented")
}
