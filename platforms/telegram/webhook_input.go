package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"time"
)

type TelegramWebhookInput struct {
	update    tgbotapi.Update
}

type TelegramWebhookUpdateProvider interface {
	TgUpdate() tgbotapi.Update
}

func (whi TelegramWebhookInput) TgUpdate() tgbotapi.Update {
	return whi.update
}

var _ bots.WebhookInput = (*TelegramWebhookTextMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookContactMessage)(nil)
var _ bots.WebhookInput = (*TelegramWebhookInlineQuery)(nil)
var _ bots.WebhookInput = (*TelegramWebhookChosenInlineResult)(nil)
var _ bots.WebhookInput = (*TelegramWebhookCallbackQuery)(nil)
var _ bots.WebhookInput = (*TelegramWebhookNewChatMembersMessage)(nil)

func (whi TelegramWebhookInput) GetID() interface{} {
	return whi.update.UpdateID
}

func NewTelegramWebhookInput(update tgbotapi.Update) bots.WebhookInput {
	input := TelegramWebhookInput{update: update}
	switch {
	case update.Message != nil:
		switch {
		case update.Message.Text != "":
			return NewTelegramWebhookTextMessage(input)
		case update.Message.Contact != nil:
			return NewTelegramWebhookContact(input)
		case update.Message.NewChatMembers != nil:
			return NewTelegramWebhookNewChatMembersMessage(input)
		default:
			return nil // TODO: Should we log it properly?
			//panic("Unexpected content of update.Message (Text is empty and no Contact)")
		}
	case update.EditedMessage != nil:
		if update.EditedMessage.Text == "" {
			panic("Unsupported EditedMessage type")
		}
		return NewTelegramWebhookTextMessage(input)
	case update.InlineQuery != nil:
		return NewTelegramWebhookInlineQuery(input)
	case update.CallbackQuery != nil:
		return NewTelegramWebhookCallbackQuery(input)
	case update.ChosenInlineResult != nil:
		return NewTelegramWebhookChosenInlineResult(input)
	default:
		return nil
	}
}

func (whi TelegramWebhookInput) GetSender() bots.WebhookSender {
	switch {
	case whi.update.Message != nil:
		return TelegramSender{tgUser: whi.update.Message.From}
	case whi.update.EditedMessage != nil:
		return TelegramSender{tgUser: whi.update.EditedMessage.From}
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

func (whi TelegramWebhookInput) GetRecipient() bots.WebhookRecipient {
	panic("Not implemented")
}

func (whi TelegramWebhookInput) GetTime() time.Time {
	if whi.update.Message != nil {
		return whi.update.Message.Time()
	}
	if whi.update.EditedMessage != nil {
		return whi.update.EditedMessage.Time()
	}
	return time.Time{}
}

func (whi TelegramWebhookInput) StringID() string {
	return ""
}

func (whi TelegramWebhookInput) TelegramChatID() int64 {
	if whi.update.Message != nil {
		return whi.update.Message.Chat.ID
	}
	if whi.update.EditedMessage != nil {
		return whi.update.EditedMessage.Chat.ID
	}
	panic("Can't get Telgram chat ID from `update.Message` or `update.EditedMessage`.")
}

func (whi TelegramWebhookInput) Chat() bots.WebhookChat {
	update := whi.update
	if update.Message != nil {
		return TelegramWebhookChat{
			chat: update.Message.Chat,
		}
	} else if update.EditedMessage != nil {
			return TelegramWebhookChat{
				chat: update.EditedMessage.Chat,
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
