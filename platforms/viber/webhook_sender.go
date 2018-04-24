package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

// ViberSender sends messages to Viber
type ViberSender struct {
	sender viberinterface.CallbackSender
}

var _ bots.WebhookSender = (*ViberSender)(nil)

// IsBotUser returns true if message sent by a bot, always false for Viber
func (ViberSender) IsBotUser() bool {
	return false
}

// GetID returns ID of Viber user
func (s ViberSender) GetID() interface{} {
	return s.sender.ID
}

// GetFirstName returns first name of Viber user
func (s ViberSender) GetFirstName() string {
	return ""
}

// GetLastName returns last name of Viber user
func (s ViberSender) GetLastName() string {
	return ""
}

// GetUserName returns nickname of Viber user
func (s ViberSender) GetUserName() string {
	return s.sender.Name
}

// GetAvatar returns avatar URL
func (s ViberSender) GetAvatar() string {
	return s.sender.Avatar
}

// Platform returns 'viber'
func (_ ViberSender) Platform() string {
	return ViberPlatformID
}

// GetLanguage is not implemented yet
func (_ ViberSender) GetLanguage() string {
	return "" // TODO: Check if we can return actual
}

func newViberSender(sender viberinterface.CallbackSender) ViberSender {
	return ViberSender{sender: sender}
}
