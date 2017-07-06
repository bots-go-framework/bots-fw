package bots

//go:generate ffjson $GOFILE

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"golang.org/x/net/context"
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-api-fbm"
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

type MessageFormat int

const (
	MessageFormatText MessageFormat = iota
	MessageFormatHTML
	MessageFormatMarkdown
)

const NoMessageToSend = "<NO_MESSAGE_TO_SEND>"

type MessageFromBot struct {
	Text                      string `json:",omitempty"`
	Format                    MessageFormat `json:",omitempty"`
	DisableWebPagePreview     bool `json:",omitempty"`
	DisableNotification       bool `json:",omitempty"`
													  //Keyboard              Keyboard
	TelegramKeyboard          interface{} `json:",omitempty"` // TODO: cast to a specific interface?
	ViberKeyboard             *viberinterface.Keyboard `json:",omitempty"`
	FbmAttachment				  *fbm_api.RequestAttachment `json:",omitempty"`
													  // TODO: One of this 2 is duplicate!?
	TelegramInlineConfig      *tgbotapi.InlineConfig `json:",omitempty"`
													  //TelegramInlineAnswer      *tgbotapi.InlineConfig
	TelegramCallbackAnswer    *tgbotapi.CallbackConfig `json:",omitempty"`
													  //
	TelegramEditMessageText   *tgbotapi.EditMessageTextConfig `json:",omitempty"`
	TelegramEditMessageMarkup *tgbotapi.EditMessageReplyMarkupConfig `json:",omitempty"`
	TelegramChatID            int64 `json:",omitempty"`
	IsReplyToInputMessage     bool `json:",omitempty"`
}

//type Keyboard interface {
//	IsKeyboard()
//}
//
//type KeyboardSelective struct {
//	Selective       bool
//}
//func (kb KeyboardSelective) IsKeyboard() {}
//
//type ForceReply struct {
//	KeyboardSelective
//	ForceReply      bool
//}
//var _ Keyboard = (*ForceReply)(nil)
//
//type ReplyKeyboardHide struct {
//	KeyboardSelective
//	HideKeyboard    bool
//}
//var _ Keyboard = (*ReplyKeyboardHide)(nil)
//
//type ReplyKeyboardMarkup struct {
//	KeyboardSelective
//	ResizeKeyboard  bool
//	OneTimeKeyboard bool
//	Buttons         [][]KeyboardButton
//}
//var _ Keyboard = (*ReplyKeyboardMarkup)(nil)
