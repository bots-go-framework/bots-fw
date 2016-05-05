package telegram_bot

import "github.com/strongo/bots-framework/core"

type TelegramWebhookActor struct {

}

var _ bots.WebhookActor = (*TelegramWebhookActor)(nil)

func (a TelegramWebhookActor) GetID() int64 {
	return 0
}