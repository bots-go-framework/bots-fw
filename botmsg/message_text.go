package botmsg

import "github.com/bots-go-framework/bots-go-core/botkb"

var _ BotMessage = (*TextMessageFromBot)(nil)

// TextMessageFromBot is a text output message from bot to user
type TextMessageFromBot struct {
	Text                  string         `json:",omitempty"`
	Format                Format         `json:",omitempty"`
	DisableWebPagePreview bool           `json:",omitempty"`
	DisableNotification   bool           `json:",omitempty"`
	Keyboard              botkb.Keyboard `json:",omitempty"`
	IsEdit                bool           `json:",omitempty"`
	EditMessageUID        MessageUID     `json:",omitempty"`
}

func (m *TextMessageFromBot) BotEndpoint() string {
	return "sendMessage"
}

// BotMessageType returns if we want to send a new message or edit existing one
func (m *TextMessageFromBot) BotMessageType() Type {
	if m.IsEdit {
		return TypeEditMessage
	}
	return TypeText
}
