package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

type ViberSender struct {
	sender viberinterface.CallbackSender
}

var _ bots.WebhookSender = (*ViberSender)(nil)

func (ViberSender) IsBotUser() bool {
	return false
}

func (s ViberSender) GetID() interface{} {
	return s.sender.ID
}

func (s ViberSender) GetFirstName() string {
	return ""
}

func (s ViberSender) GetLastName() string {
	return ""
}

func (s ViberSender) GetUserName() string {
	return s.sender.Name
}

func (s ViberSender) GetAvatar() string {
	return s.sender.Avatar
}

func (_ ViberSender) Platform() string {
	return ViberPlatformID
}

func (_ ViberSender) GetLanguage() string {
	return "" // TODO: Check if we can return actual
}

func newViberSender(sender viberinterface.CallbackSender) ViberSender {
	return ViberSender{sender: sender}
}
