package botkb

type ButtonType int

type Button interface {
	GetText() string
	ButtonType() ButtonType
}

const (
	ButtonTypeData ButtonType = iota
	ButtonTypeURL
	ButtonTypeSwitchInlineQuery
	ButtonTypeSwitchInlineQueryCurrentChat
)

var _ Button = (*DataButton)(nil)
var _ Button = (*UrlButton)(nil)
var _ Button = (*SwitchInlineQueryButton)(nil)
var _ Button = (*SwitchInlineQueryCurrentChatButton)(nil)

func NewDataButton(text, data string) *DataButton {
	return &DataButton{Text: text, Data: data}
}

type DataButton struct {
	Text string
	Data string
}

func (b DataButton) GetText() string {
	return b.Text
}

func (DataButton) ButtonType() ButtonType {
	return ButtonTypeData
}

type UrlButton struct {
	Text string
	URL  string
}

func NewUrlButton(text, url string) *UrlButton {
	return &UrlButton{Text: text, URL: url}
}

func (UrlButton) ButtonType() ButtonType {
	return ButtonTypeURL
}

func (b UrlButton) GetText() string {
	return b.Text
}

type SwitchInlineQueryButton struct {
	Text  string
	Query string
}

func NewSwitchInlineQueryButton(text, query string) *SwitchInlineQueryButton {
	return &SwitchInlineQueryButton{Text: text, Query: query}
}

func (SwitchInlineQueryButton) ButtonType() ButtonType {
	return ButtonTypeSwitchInlineQuery
}

func (b SwitchInlineQueryButton) GetText() string {
	return b.Text
}

type SwitchInlineQueryCurrentChatButton struct {
	Text  string
	Query string
}

func NewSwitchInlineQueryCurrentChatButton(text, query string) *SwitchInlineQueryCurrentChatButton {
	return &SwitchInlineQueryCurrentChatButton{Text: text, Query: query}
}

func (SwitchInlineQueryCurrentChatButton) ButtonType() ButtonType {
	return ButtonTypeSwitchInlineQueryCurrentChat
}

func (b SwitchInlineQueryCurrentChatButton) GetText() string {
	return b.Text
}
