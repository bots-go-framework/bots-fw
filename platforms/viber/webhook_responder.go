package viber

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

// viberWebhookResponder is wrapper for Viber API
type viberWebhookResponder struct {
	//w   http.ResponseWriter
	whc *viberWebhookContext
}

var _ bots.WebhookResponder = (*viberWebhookResponder)(nil)

func newViberWebhookResponder(whc *viberWebhookContext) viberWebhookResponder {
	responder := viberWebhookResponder{whc: whc} // We need a dedicated to get rid of type assertion
	whc.responder = responder
	return responder
}

func (r viberWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotAPISendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	log.Debugf(c, "viberWebhookResponder.SendMessage()...")
	botSettings := r.whc.GetBotSettings()
	viberBotAPI := viberbotapi.NewViberBotApiWithHttpClient(botSettings.Token, r.whc.BotContext.BotHost.GetHTTPClient(c))
	log.Debugf(c, "Keyboard: %v", m.Keyboard)

	var viberKeyboard *viberinterface.Keyboard
	if viberKeyboard, ok := m.Keyboard.(*viberinterface.Keyboard); ok && viberKeyboard != nil {
		viberKeyboard.Type = "keyboard"
	}

	textMessage := viberinterface.NewTextMessage(r.whc.getViberSenderID(), "track-data", m.Text, viberKeyboard)
	requestBody, response, err := viberBotAPI.SendMessage(textMessage)
	if err != nil {
		err = errors.Wrap(err, "Failed to send message to Viber")
		log.Errorf(c, err.Error())
	}

	log.Debugf(c, "Request body: %v", (string)(requestBody))
	if response.Status == 0 {
		log.Debugf(c, "Successfully sent to Viber")
	} else {
		switch response.Status { // https://developers.viber.com/customer/en/portal/articles/2541337-error-codes?b_id=15145
		case 2:
			log.Errorf(c, "Viber response.Status=%v: %v: [%v]", response.Status, response.StatusMessage, botSettings.Token)
		default:
			log.Errorf(c, "Viber response.Status=%v: %v", response.Status, response.StatusMessage)
		}
	}
	return resp, err
}
