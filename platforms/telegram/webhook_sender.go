package telegram_bot

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type TelegramSender struct {
	tgUser *tgbotapi.User
}

var _ bots.WebhookSender = (*TelegramSender)(nil)

func (s TelegramSender) GetID() interface{} {
	return s.tgUser.ID
}

func (s TelegramSender) GetFirstName() string {
	return s.tgUser.FirstName
}

func (s TelegramSender) GetLastName() string {
	return s.tgUser.LastName
}

func (s TelegramSender) GetUserName() string {
	return s.tgUser.UserName
}
