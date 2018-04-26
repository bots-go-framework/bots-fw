package telegram

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

type tgSender struct {
	tgUser *tgbotapi.User
}

func (tgSender) IsBotUser() bool { // TODO: Can we get rid of it here?
	return false
}

var _ bots.WebhookSender = (*tgSender)(nil)

func (s tgSender) GetID() interface{} {
	return s.tgUser.ID
}

func (s tgSender) GetFirstName() string {
	return s.tgUser.FirstName
}

func (s tgSender) GetLastName() string {
	return s.tgUser.LastName
}

func (s tgSender) GetUserName() string {
	return s.tgUser.UserName
}

func (tgSender) Platform() string {
	return PlatformID
}

func (tgSender) GetAvatar() string {
	return ""
}

func (s tgSender) GetLanguage() string {
	return s.tgUser.LanguageCode
}
