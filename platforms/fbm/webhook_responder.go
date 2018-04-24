package fbm_bot

import (
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine/urlfetch"
)

type FbmWebhookResponder struct {
	whc *FbmWebhookContext
}

var _ bots.WebhookResponder = (*FbmWebhookResponder)(nil)

func NewFbmWebhookResponder(whc *FbmWebhookContext) FbmWebhookResponder {
	responder := FbmWebhookResponder{whc: whc} // We need a dedicated to get rid of type assertion
	whc.responder = responder
	return responder
}

func (r FbmWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotApiSendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	log.Debugf(c, "FbmWebhookResponder.SendMessage()...")

	if m.Text != "" && m.FbmAttachment != nil {
		err = errors.New("m.Text is empty string && m.FbmAttachment != nil")
		return
	}

	request := fbm_api.Request{
		NotificationType: fbm_api.RequestNotificationTypeNoPush,
		//Recipient: fbm_api.RequestRecipient{},
		Message: fbm_api.RequestMessage{
			Text:       m.Text,
			Attachment: m.FbmAttachment,
		},
	}

	if request.Recipient.Id, err = r.whc.BotChatID(); err != nil {
		err = errors.WithMessage(err, "failed to call r.whc.BotChatID()")
		return
	} else if request.Recipient.Id == "" {
		err = errors.New("Unknown recipient as r.whc.BotChatID() returned an empty string")
		return
	}

	graphApi := fbm_api.NewGraphApi(urlfetch.Client(c), r.whc.GetBotSettings().Token)

	if err = graphApi.SendMessage(c, request); err != nil {
		return
	}

	return
}
