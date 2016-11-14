package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

type ViberSender struct {
	sender viberinterface.CallbackSender
}

var _ bots.WebhookSender = (*ViberSender)(nil)

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

func (_ ViberSender) Platform() string {
	return "viber"
}

func newViberSender(sender viberinterface.CallbackSender) ViberSender {
	return ViberSender{sender: sender}
}