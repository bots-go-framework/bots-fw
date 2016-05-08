package bots

import "net/http"

type EntryInputs struct {
	Entry  WebhookEntry
	Inputs []WebhookInput
}

type EntryInput struct {
	Entry WebhookEntry
	Input WebhookInput
}

type TranslatorProvider func(logger Logger) Translator

type BaseHandler struct {
	WebhookDriver
	BotHost
	BotPlatform
	TranslatorProvider TranslatorProvider
}

func (h BaseHandler) GetTranslator(r *http.Request) Translator {
	return h.TranslatorProvider(h.BotHost.GetLogger(r))
}

type MessageFormat int

const (
	MessageFormatText MessageFormat = iota
	MessageFormatHTML
	MessageFormatMarkdown
)

type MessageFromBot struct {
	Text                  string
	Format                MessageFormat
	DisableWebPagePreview bool
	Keyboard              Keyboard
	IsReplyToInputMessage bool
}

type Keyboard struct {
	HideKeyboard    bool
	ResizeKeyboard  bool
	ForceReply      bool
	Selective       bool
	OneTimeKeyboard bool
	Buttons         [][]string
}
