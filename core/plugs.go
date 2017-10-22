package bots

import (
	"net/url"
)

var IgnoreCommand = Command{
	Code: "bots.IgnoreCommand",
	Action: func(_ WebhookContext) (m MessageFromBot, err error) {
		return
	},
	CallbackAction: func(_ WebhookContext, _ *url.URL) (m MessageFromBot, err error) {
		return
	},
}
