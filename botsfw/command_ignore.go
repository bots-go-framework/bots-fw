package botsfw

import (
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"net/url"
)

// IgnoreCommand is a command that does nothing
var IgnoreCommand = Command{
	Code: "bots.IgnoreCommand",
	Action: func(_ WebhookContext) (m botmsg.MessageFromBot, err error) {
		return
	},
	CallbackAction: func(_ WebhookContext, _ *url.URL) (m botmsg.MessageFromBot, err error) {
		return
	},
	TextAction: func(_ WebhookContext, _ string) (m botmsg.MessageFromBot, err error) {
		return
	},
	InlineQueryAction: func(_ WebhookContext, _ botinput.InlineQuery, _ *url.URL) (m botmsg.MessageFromBot, err error) {
		return
	},
	ChosenInlineResultAction: func(_ WebhookContext, _ botinput.ChosenInlineResult, _ *url.URL) (m botmsg.MessageFromBot, err error) {
		return
	},
}
