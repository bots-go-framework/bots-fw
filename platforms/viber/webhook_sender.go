package viber

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

// Sender sends messages to Viber
type Sender struct {
	sender viberinterface.CallbackSender
}

var _ bots.WebhookSender = (*Sender)(nil)

// IsBotUser returns true if message sent by a bot, always false for Viber
func (Sender) IsBotUser() bool {
	return false
}

// GetID returns ID of Viber user
func (s Sender) GetID() interface{} {
	return s.sender.ID
}

// GetFirstName returns first name of Viber user
func (s Sender) GetFirstName() string {
	return ""
}

// GetLastName returns last name of Viber user
func (s Sender) GetLastName() string {
	return ""
}

// GetUserName returns nickname of Viber user
func (s Sender) GetUserName() string {
	return s.sender.Name
}

// GetAvatar returns avatar URL
func (s Sender) GetAvatar() string {
	return s.sender.Avatar
}

// Platform returns 'viber'
func (Sender) Platform() string {
	return PlatformID
}

// GetLanguage is not implemented yet
func (Sender) GetLanguage() string {
	return "" // TODO: Check if we can return actual
}

func newViberSender(sender viberinterface.CallbackSender) Sender {
	return Sender{sender: sender}
}
