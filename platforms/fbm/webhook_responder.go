package fbm_bot

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"github.com/strongo/bots-api-fbm"
	"bytes"
	"io/ioutil"
	"github.com/pkg/errors"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app/log"
)

type FbmWebhookResponder struct {
	//w   http.ResponseWriter
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

	//fbmWhc := (FbmWebhookContext{})(r.whc)

	request := fbm_api.Request{
		NotificationType: fbm_api.RequestNotificationTypeNoPush,
		Recipient: fbm_api.RequestRecipient{Id: r.whc.BotChatID()},
		Message: fbm_api.RequestMessage{
			Text: m.Text,
			Attachment: m.FbmAttachment,
		},
	}

	data, err := ffjson.MarshalFast(request)
	if err != nil {
		err = errors.Wrap(err, "Failed to marshal request to JSON")
		return
	}

	accessToken := r.whc.GetBotSettings().Token
	log.Debugf(c, "Posting to FB Messenger API (accessToken=%v):\n%v", accessToken, string(data))

	httpClient := urlfetch.Client(c)
	apiResponse, err := httpClient.Post(
		"https://graph.facebook.com/v2.6/me/messages?access_token=" + accessToken,
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		err = errors.Wrap(err, "Failed to post to FB Messenger API")
		return
	}

	if apiResponse.Body != nil {
		defer apiResponse.Body.Close()
	}

	respData, err2 := ioutil.ReadAll(apiResponse.Body)
	if err2 != nil {
		err = errors.Wrap(err2, "Failed to read response body")
		return
	}
	switch apiResponse.StatusCode {
	case http.StatusBadRequest:
		err = errors.New(fmt.Sprintf("Bad request: %v", string(respData)))
		return
	}
	log.Debugf(c, "Gor from response FB Messenger API status=%v: %v", apiResponse.Status, string(respData))
	return
}
