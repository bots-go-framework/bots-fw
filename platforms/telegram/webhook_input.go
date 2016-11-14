package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"time"
)

type telegramWebhookInput struct {
	update    tgbotapi.Update
}

type TelegramWebhookUpdateProvider interface {
	TgUpdate() tgbotapi.Update
}

func (i telegramWebhookInput) TgUpdate() tgbotapi.Update {
	return i.update
}

var _ bots.WebhookInput = (*TelegramWebhookTextMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookContactMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookInlineQuery)(nil)
var _ bots.WebhookInput = (*TelegramWebhookChosenInlineResult)(nil)
var _ bots.WebhookInput = (*TelegramWebhookCallbackQuery)(nil)

func (input telegramWebhookInput) GetID() interface{} {
	return input.update.UpdateID
}

func NewTelegramWebhookInput(update tgbotapi.Update) bots.WebhookInput {
	input := telegramWebhookInput{update: update}
	switch {
	case update.Message != nil:
		switch {
		case update.Message.Text != "":
			return NewTelegramWebhookTextMessage(input)
		case update.Message.Contact != nil:
			return NewTelegramWebhookContact(input)
		default:
			panic("Unexpected content of update.Message (Text is empty and no Contact)")
		}
	case update.InlineQuery != nil:
		return NewTelegramWebhookInlineQuery(input)
	case update.CallbackQuery != nil:
		return NewTelegramWebhookCallbackQuery(input)
	case update.ChosenInlineResult != nil:
		return NewTelegramWebhookChosenInlineResult(input)
	default:
		panic("Unkonwn content of Telegram update message")
	}
}

func (whi telegramWebhookInput) GetSender() bots.WebhookSender {
	switch {
	case whi.update.Message != nil:
		return TelegramSender{tgUser: whi.update.Message.From}
	case whi.update.CallbackQuery != nil:
		return TelegramSender{tgUser: whi.update.CallbackQuery.From}
	case whi.update.InlineQuery != nil:
		return TelegramSender{tgUser: whi.update.InlineQuery.From}
	case whi.update.ChosenInlineResult != nil:
		return TelegramSender{tgUser: whi.update.ChosenInlineResult.From}
	default:
		panic("Unknown From sender")
	}
}

func (whi telegramWebhookInput) GetRecipient() bots.WebhookRecipient {
	panic("Not implemented")
}

func (whi telegramWebhookInput) GetTime() time.Time {
	if whi.update.Message != nil {
		return whi.update.Message.Time()
	}
	return time.Time{}
}

func (whi telegramWebhookInput) StringID() string {
	return ""
}

func (whi telegramWebhookInput) Chat() bots.WebhookChat {
	update := whi.update
	if update.Message != nil {
		return TelegramWebhookChat{
			chat: update.Message.Chat,
		}
	} else {
		callbackQuery := update.CallbackQuery
		if callbackQuery != nil && callbackQuery.Message != nil {
			return TelegramWebhookChat{
				chat: callbackQuery.Message.Chat,
			}
		}
	}
	return nil
}
