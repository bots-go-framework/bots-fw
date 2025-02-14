package botsfw

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botinput"
	"net/url"
)

// CommandAction defines an action bot can perform in response to a command
type CommandAction func(whc WebhookContext) (m MessageFromBot, err error)

type TextAction func(whc WebhookContext, text string) (m MessageFromBot, err error)

// CallbackAction defines a callback action bot can perform in response to a callback command
type CallbackAction func(whc WebhookContext, callbackUrl *url.URL) (m MessageFromBot, err error)

type InlineQueryAction func(whc WebhookContext, inlineQuery botinput.WebhookInlineQuery, queryUrl *url.URL) (m MessageFromBot, err error)

type ChosenInlineResultAction func(whc WebhookContext, chosenResult botinput.WebhookChosenInlineResult, queryUrl *url.URL) (m MessageFromBot, err error)

// CommandMatcher returns true if action is matched to user input
type CommandMatcher func(command Command, whc WebhookContext) bool

// DefaultTitle key
const DefaultTitle = "" //

// ShortTitle key
const ShortTitle = "short_title"

//const LongTitle = "long_title"

type CommandCode string

// Command defines command metadata and action
type Command struct {
	Code       CommandCode
	InputTypes []botinput.WebhookInputType // Instant match if != WebhookInputUnknown && == whc.InputTypes()
	Icon       string
	Replies    []Command
	Title      string
	Titles     map[string]string
	ExactMatch string
	Commands   []string
	Matcher    CommandMatcher
	//
	Action                   CommandAction
	TextAction               TextAction
	CallbackAction           CallbackAction
	InlineQueryAction        InlineQueryAction
	ChosenInlineResultAction ChosenInlineResultAction
}

//goland:noinspection GoUnusedExportedFunction
func NewInlineQueryCommand(code CommandCode, action CommandAction) Command {
	return Command{
		Code:       code,
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputInlineQuery},
		Action:     action,
	}
}

// NewCallbackCommand create a definition of a callback command
//
//goland:noinspection GoUnusedExportedFunction
func NewCallbackCommand(code CommandCode, action CallbackAction) Command {
	return Command{
		Code:           code,
		InputTypes:     []botinput.WebhookInputType{botinput.WebhookInputCallbackQuery},
		Commands:       []string{"/" + string(code)},
		CallbackAction: action,
	}
}

func (c Command) String() string {
	return fmt.Sprintf("Command{Code: '%v', InputTypes: %v, Icon: '%v', Title: '%v', ExactMatch: '%v', len(Command): %v, len(Replies): %v}", c.Code, c.InputTypes, c.Icon, c.Title, c.ExactMatch, len(c.Commands), len(c.Replies))
}

// CommandTextNoTrans returns a title for a command (pre-translated)
func CommandTextNoTrans(title, icon string) string {
	if title == "" && icon != "" {
		return icon
	} else if title != "" && icon == "" {
		return title
	} else if title != "" && icon != "" {
		return title + " " + icon
	}
	return "<NO_TITLE_OR_ICON>"
}

// DefaultTitle returns a default title for a command in current Locale
func (c Command) DefaultTitle(whc WebhookContext) string {
	return c.TitleByKey(DefaultTitle, whc)
}

// TitleByKey returns a short/long title for a command in current Locale
func (c Command) TitleByKey(key string, whc WebhookContext) string {
	var title string
	if key == DefaultTitle && c.Title != "" {
		title = c.Title
	} else if val, ok := c.Titles[key]; ok {
		title = val
	}

	if c.Icon == "" {
		if title == "" {
			title = string(c.Code)
		} else {
			title = whc.Translate(title)
		}
	} else {
		if title == "" {
			title = CommandTextNoTrans("", c.Icon)
		} else {
			title = whc.CommandText(title, c.Icon)
		}
	}
	return title
}
