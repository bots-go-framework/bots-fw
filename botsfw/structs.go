package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-go-core/botkb"
	"github.com/strongo/i18n"
	"strconv"
	//"github.com/strongo/bots-api-fbm"
)

// EntryInputs provides information on parsed inputs from bot API request
type EntryInputs struct {
	Entry  botinput.WebhookEntry
	Inputs []botinput.WebhookInput
}

// EntryInput provides information on parsed input from bot API request
type EntryInput struct {
	Entry botinput.WebhookEntry
	Input botinput.WebhookInput
}

// TranslatorProvider translates texts
type TranslatorProvider func(c context.Context) i18n.Translator

// WebhookHandlerBase provides base implementation for a bot handler
type WebhookHandlerBase struct {
	WebhookDriver
	BotHost
	BotPlatform
	//RecordsMaker        botsfwmodels.BotRecordsMaker
	RecordsFieldsSetter BotRecordsFieldsSetter
	TranslatorProvider  TranslatorProvider
	//DataAccess          botsfwdal.DataAccess
}

// Register driver
func (bh *WebhookHandlerBase) Register(d WebhookDriver, h BotHost) {
	if d == nil {
		panic("WebhookDriver == nil")
	}
	if h == nil {
		panic("BotHost == nil")
	}
	bh.WebhookDriver = d
	bh.BotHost = h
}

// MessageFormat specify formatting of a text message to BOT (e.g. Text, HTML, MarkDown)
type MessageFormat int

//goland:noinspection GoUnusedConst
const (
	// MessageFormatText is for text messages
	MessageFormatText MessageFormat = iota
	// MessageFormatHTML is for HTML messages
	MessageFormatHTML
	// MessageFormatMarkdown is for markdown messages
	MessageFormatMarkdown
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

// AttachmentType to a bot message
type AttachmentType int

//goland:noinspection GoUnusedConst
const (
	// AttachmentTypeNone says there is no attachment
	AttachmentTypeNone AttachmentType = iota

	// AttachmentTypeAudio is for audio attachments
	AttachmentTypeAudio

	// AttachmentTypeFile is for file attachments
	AttachmentTypeFile

	// AttachmentTypeImage is for image attachments
	AttachmentTypeImage

	// AttachmentTypeVideo is for video attachments
	AttachmentTypeVideo
)

// Attachment to a bot message
type Attachment interface {
	AttachmentType() AttachmentType
}

// BotMessageType defines type of an output message from bot to user
type BotMessageType int

const (
	// BotMessageTypeUndefined unknown type
	BotMessageTypeUndefined BotMessageType = iota
	// BotMessageTypeCallbackAnswer sends a callback answer
	BotMessageTypeCallbackAnswer
	// BotMessageTypeInlineResults sends inline results
	BotMessageTypeInlineResults
	// BotMessageTypeText sends text reply
	BotMessageTypeText
	// BotMessageTypeEditMessage edit previously sent message
	BotMessageTypeEditMessage
	// BotMessageTypeLeaveChat commands messenger to kick off bot from a botChat
	BotMessageTypeLeaveChat
	// BotMessageTypeExportChatInviteLink sends invite link
	BotMessageTypeExportChatInviteLink

	BotMessageTypeSendPhoto

	BotMessageTypeSendInvoice
	BotMessageTypeCreateInvoiceLink
	BotMessageTypeAnswerPreCheckoutQuery

	BotMessageTypeSetDescription
	BotMessageTypeSetShortDescription
	BotMessageTypeSetCommands
)

// BotMessage is an output message from bot to user
type BotMessage interface {
	BotMessageType() BotMessageType
	//BotEndpoint() string
}

// TextMessageFromBot is a text output message from bot to user
type TextMessageFromBot struct {
	Text                  string         `json:",omitempty"`
	Format                MessageFormat  `json:",omitempty"`
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
func (m *TextMessageFromBot) BotMessageType() BotMessageType {
	if m.IsEdit {
		return BotMessageTypeEditMessage
	}
	return BotMessageTypeText
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
