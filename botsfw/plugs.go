package botsfw

import (
	"github.com/bots-go-framework/bots-fw/botinput"
	"net/url"
)

// IgnoreCommand is a command that does nothing
var IgnoreCommand = Command{
	Code: "bots.IgnoreCommand",
	Action: func(_ WebhookContext) (m MessageFromBot, err error) {
		return
	},
	CallbackAction: func(_ WebhookContext, _ *url.URL) (m MessageFromBot, err error) {
		return
	},
	TextAction: func(_ WebhookContext, _ string) (m MessageFromBot, err error) {
		return
	},
	InlineQueryAction: func(_ WebhookContext, _ botinput.WebhookInlineQuery, _ *url.URL) (m MessageFromBot, err error) {
		return
	},
	ChosenInlineResultAction: func(_ WebhookContext, _ botinput.WebhookChosenInlineResult, _ *url.URL) (m MessageFromBot, err error) {
		return
	},
}
