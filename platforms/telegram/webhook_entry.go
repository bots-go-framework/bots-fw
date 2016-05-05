package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"time"
	"github.com/strongo/bots-api-telegram"
)

type TelegramWebhookEntry struct {
	update tgbotapi.Update
}

var _ bots.WebhookEntry = (*TelegramWebhookEntry)(nil)

func (e TelegramWebhookEntry) GetID() int64 {
	return (int64)(e.update.UpdateID)
}

func (e TelegramWebhookEntry) GetTime() time.Time {
	return e.update.Message.Time()
}