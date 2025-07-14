package botmsg

import (
	"github.com/bots-go-framework/bots-fw/botkb"
	"strconv"
)

// BotAPISendMessageChannel specifies messenger channel
type BotAPISendMessageChannel string

// Format specifies formatting of a text message to BOT (e.g. TypeText, HTML, MarkDown)
type Format int

//goland:noinspection GoUnusedConst
const (
	// FormatText is for text messages
	FormatText Format = iota
	// FormatHTML is for HTML messages
	FormatHTML
	// FormatMarkdown is for markdown messages
	FormatMarkdown
)

// NoMessageToSend returned explicitly if we don't want to reply to user intput
const NoMessageToSend = "<NO_MESSAGE_TO_SEND>"

// ChatUID returns botChat ID as unique string
type ChatUID interface {
	ChatUID() string
}

// ChatIntID returns botChat ID as unique integer
type ChatIntID int64

// ChatUID returns botChat ID as unique string
func (chatUID ChatIntID) ChatUID() string {
	return strconv.FormatInt(int64(chatUID), 10)
}

// MessageUID is unique message ID as string
type MessageUID interface {
	UID() string
}

// BotMessage is an output message from bot to user
type BotMessage interface {
	BotMessageType() Type
	//BotEndpoint() string
}

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

// Type returns if we want to send a new message or edit existing one
func (m *TextMessageFromBot) BotMessageType() Type {
	if m.IsEdit {
		return TypeEditMessage
	}
	return TypeText
}

var _ BotMessage = (*TextMessageFromBot)(nil)

// MessageFromBot keeps all the details of answer from bot to user
//
//goland:noinspection GoDeprecation
type MessageFromBot struct {
	ResponseChannel BotAPISendMessageChannel `json:"-,omitempty"` // For debug purposes
	ToChat          ChatUID                  `json:",omitempty"`

	// To be used with Telegram to edit an arbitrary message.
	// Do not use this field directly when you want to edit the callback message
	EditMessageIntID int `json:"editMessageIntID,omitempty"`

	// This is a shortcut to MessageFromBot{}.BotMessage = TextMessageFromBot{text: "abc"}
	TextMessageFromBot // TODO: This feels wrong and need to be refactored! Use BotMessage instead

	BotMessage BotMessage `json:",omitempty"`
	//FbmAttachment      *fbmbotapi.RequestAttachment `json:",omitempty"` // deprecated
}
