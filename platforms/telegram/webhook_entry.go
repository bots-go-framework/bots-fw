package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"time"
)

type TelegramWebhookEntry struct {
	update *tgbotapi.Update
}

var _ bots.WebhookEntry = (*TelegramWebhookEntry)(nil)

func (e TelegramWebhookEntry) GetID() interface{} {
	return e.update.UpdateID
}

func (e TelegramWebhookEntry) GetTime() time.Time {
	if e.update.Message != nil {
		return e.update.Message.Time()
	}
	if e.update.EditedMessage != nil {
		return e.update.EditedMessage.Time()
	}
	panic("Both `update.Message` & `update.EditedMessage` are nil.")
}
