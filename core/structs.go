package bots

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
)

type EntryInputs struct {
	Entry  WebhookEntry
	Inputs []WebhookInput
}

type EntryInput struct {
	Entry WebhookEntry
	Input WebhookInput
}

type TranslatorProvider func(logger strongo.Logger) strongo.Translator

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
	Text                  string
	Format                MessageFormat
	DisableWebPagePreview bool
	//Keyboard              Keyboard
	TelegramKeyboard          interface{}
	TelegramInlineAnswer      *tgbotapi.InlineConfig
	TelegramEditMessageText   *tgbotapi.EditMessageTextConfig
	TelegramEditMessageMarkup *tgbotapi.EditMessageReplyMarkupConfig
	TelegramInlineConfig      *tgbotapi.InlineConfig
	TelegramChatID            int64
	IsReplyToInputMessage     bool
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
