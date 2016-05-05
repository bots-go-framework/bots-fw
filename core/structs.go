package bots

type EntryInputs struct {
	Entry  WebhookEntry
	Inputs []WebhookInput
}

type EntryInput struct {
	Entry WebhookEntry
	Input WebhookInput
}

type BaseHandler struct {
	WebhookDriver
	BotHost
	BotPlatform
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
