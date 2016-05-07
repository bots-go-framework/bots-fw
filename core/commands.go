package bots

type CommandAction func(WebhookContext) (MessageFromBot, error)

type CommandMatcher func(Command, WebhookContext) bool

const DEFAULT_TITLE = ""
const SHORT_TITLE = "short_title"

//const LONG_TITLE = "long_title"

type Command struct {
	Icon       string
	Replies    []Command
	Code       string
	Title      string
	Titles     map[string]string
	ExactMatch string
	Commands   []string
	Matcher    CommandMatcher
	Action     CommandAction
}

func (whcb *WebhookContextBase) CommandTitle(title, icon string) string {
	title = whcb.Translate(title)
	return CommandTitleNoTrans(title, icon)
}

func CommandTitleNoTrans(title, icon string) string {
	if title == "" && icon != "" {
		return icon
	} else if title != "" && icon == "" {
		return title
	} else if title != "" && icon != "" {
		return title + " " + icon
	} else {
		return "<NO_TITLE_OR_ICON>"
	}
}

func (c Command) DefaultTitle(whc WebhookContext) string {
	return c.TitleByKey(DEFAULT_TITLE, whc)
}

func (c Command) TitleByKey(key string, whc WebhookContext) string {
	var title string
	if key == DEFAULT_TITLE && c.Title != "" {
		title = c.Title
	} else if val, ok := c.Titles[key]; ok {
		title = val
	}

	if c.Icon == "" {
		if title == "" {
			title = c.Code
		} else {
			title = whc.Translate(title)
		}
	} else {
		if title == "" {
			title = CommandTitleNoTrans("", c.Icon)
		} else {
			title = whc.CommandTitle(title, c.Icon)
		}
	}
	return title
}
