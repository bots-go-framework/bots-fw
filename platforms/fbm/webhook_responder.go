package fbm

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"google.golang.org/appengine/urlfetch"
)

// webhookResponder responds to FB Messenger
type webhookResponder struct {
	whc *fbmWebhookContext
}

var _ bots.WebhookResponder = (*webhookResponder)(nil)

// newFbmWebhookResponder creates new responder to FBM
func newFbmWebhookResponder(whc *fbmWebhookContext) webhookResponder {
	responder := webhookResponder{whc: whc} // We need a dedicated to get rid of type assertion
	whc.responder = responder
	return responder
}

// SendMessage sends message to FBM
func (r webhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotAPISendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	log.Debugf(c, "webhookResponder.SendMessage()...")

	if m.Text != "" && m.FbmAttachment != nil {
		err = errors.New("m.Text is empty string && m.FbmAttachment != nil")
		return
	}

	request := fbmbotapi.Request{
		NotificationType: fbmbotapi.RequestNotificationTypeNoPush,
		//Recipient: fbm_api.RequestRecipient{},
		Message: fbmbotapi.RequestMessage{
			Text:       m.Text,
			Attachment: m.FbmAttachment,
		},
	}

	if request.Recipient.ID, err = r.whc.BotChatID(); err != nil {
		err = errors.WithMessage(err, "failed to call r.whc.BotChatID()")
		return
	} else if request.Recipient.ID == "" {
		err = errors.New("Unknown recipient as r.whc.BotChatID() returned an empty string")
		return
	}

	graphAPI := fbmbotapi.NewGraphAPI(urlfetch.Client(c), r.whc.GetBotSettings().Token)

	if err = graphAPI.SendMessage(c, request); err != nil {
		return
	}

	return
}
