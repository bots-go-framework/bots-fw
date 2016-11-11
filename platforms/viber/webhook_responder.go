package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"net/http"
)

type ViberWebhookResponder struct {
	w   http.ResponseWriter
	whc *ViberWebhookContext
}

var _ bots.WebhookResponder = (*ViberWebhookResponder)(nil)

func NewViberWebhookResponder(w http.ResponseWriter, whc *ViberWebhookContext) ViberWebhookResponder {
	responder := ViberWebhookResponder{w: w, whc: whc}
	whc.responder = responder
	return responder
}

func (r ViberWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotApiSendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	return
}
