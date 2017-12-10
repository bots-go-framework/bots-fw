package bots

//go:generate ffjson $GOFILE

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-api-fbm"
	"golang.org/x/net/context"
	"strconv"
)

type EntryInputs struct {
	Entry  WebhookEntry
	Inputs []WebhookInput
}

type EntryInput struct {
	Entry WebhookEntry
	Input WebhookInput
}

type TranslatorProvider func(c context.Context) strongo.Translator

type BaseHandler struct {
	WebhookDriver
	BotHost
	BotPlatform
	TranslatorProvider TranslatorProvider
}

func (bh *BaseHandler) Register(d WebhookDriver, h BotHost) {
	if d == nil {
		panic("WebhookDriver == nil")
	}
	if h == nil {
		panic("BotHost == nil")
	}
	bh.WebhookDriver = d
	bh.BotHost = h
}

type MessageFormat int

const (
	MessageFormatText MessageFormat = iota
	MessageFormatHTML
	MessageFormatMarkdown
)

const NoMessageToSend = "<NO_MESSAGE_TO_SEND>"

type ChatUID interface {
	ChatUID() string
}

type ChatIntID int64

func (chatUID ChatIntID) ChatUID() string {
	return strconv.FormatInt(int64(chatUID), 10)
}

type MessageUID interface {
	UID() string
}

type KeyboardType int

const (
	KeyboardTypeNone KeyboardType = iota
	KeyboardTypeHide
	KeyboardTypeInline
	KeyboardTypeBottom
	KeyboardTypeForceReply
)

type Keyboard interface {
	KeyboardType() KeyboardType
}

type AttachmentType int

const (
	AttachmentTypeNone AttachmentType = iota
	AttachmentTypeAudio
	AttachmentTypeFile
	AttachmentTypeImage
	AttachmentTypeVideo
)

type Attachment interface {
	AttachmentType() AttachmentType
}

type BotMessageType int

const (
	BotMessageTypeUndefined BotMessageType = iota
	BotMessageTypeCallbackAnswer
	BotMessageTypeInlineResults
	BotMessageTypeText
	BotMessageTypeEditMessage
	BotMessageTypeLeaveChat
	BotMessageTypeExportChatInviteLink
)

type BotMessage interface {
	BotMessageType() BotMessageType
}

type TextMessageFromBot struct {
	Text                  string        `json:",omitempty"`
	Format                MessageFormat `json:",omitempty"`
	DisableWebPagePreview bool          `json:",omitempty"`
	DisableNotification   bool          `json:",omitempty"`
	Keyboard              Keyboard      `json:",omitempty"`
	IsEdit                bool          `json:",omitempty"`
	EditMessageUID        MessageUID    `json:",omitempty"`
}

func (m TextMessageFromBot) BotMessageType() BotMessageType {
	if m.IsEdit {
		return BotMessageTypeEditMessage
	}
	return BotMessageTypeText
}

var _ BotMessage = (*TextMessageFromBot)(nil)

type MessageFromBot struct {
	ToChat             ChatUID                    `json:",omitempty"`
	TextMessageFromBot                            // This is a shortcut to MessageFromBot{}.BotMessage = TextMessageFromBot{text: "abc"}
	BotMessage         BotMessage                 `json:",omitempty"`
	FbmAttachment      *fbm_api.RequestAttachment `json:",omitempty"` // deprecated
}
