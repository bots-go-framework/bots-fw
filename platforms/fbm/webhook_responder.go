package fbm_bot

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"github.com/strongo/bots-api-fbm"
	"google.golang.org/appengine/urlfetch"
	"github.com/strongo/app/log"
	"github.com/pkg/errors"
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
		Recipient: fbm_api.RequestRecipient{Id: r.whc.BotChatID()},
		Message: fbm_api.RequestMessage{
			Text: m.Text,
			Attachment: m.FbmAttachment,
		},
	}

	graphApi := fbm_api.NewGraphApi(urlfetch.Client(c), r.whc.GetBotSettings().Token)

	if err = graphApi.SendMessage(c, request); err != nil {
		return
	}

	return
}
