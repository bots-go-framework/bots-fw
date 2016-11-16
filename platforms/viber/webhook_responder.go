package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"github.com/strongo/bots-api-viber"
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/pkg/errors"
	"google.golang.org/appengine/log"
)

type ViberWebhookResponder struct {
	//w   http.ResponseWriter
	whc *ViberWebhookContext
}

var _ bots.WebhookResponder = (*ViberWebhookResponder)(nil)

func NewViberWebhookResponder(whc *ViberWebhookContext) ViberWebhookResponder {
	responder := ViberWebhookResponder{whc: whc} // We need a dedicated to get rid of type assertion
	whc.responder = responder
	return responder
}

func (r ViberWebhookResponder) SendMessage(c context.Context, m bots.MessageFromBot, channel bots.BotApiSendMessageChannel) (resp bots.OnMessageSentResponse, err error) {
	logger := r.whc.Logger()
	logger.Debugf(c, "ViberWebhookResponder.SendMessage()...")
	botSettings := r.whc.GetBotSettings()
	viberBotApi := viberbotapi.NewViberBotApiWithHttpClient(botSettings.Token, r.whc.GetHttpClient())
	logger.Debugf(c, "ViberKeyboard: %v", m.ViberKeyboard)
	if m.ViberKeyboard != nil {
		m.ViberKeyboard.Type = "keyboard"
	}
	textMessage := viberinterface.NewTextMessage(r.whc.getViberSenderID(), "track-data", m.Text, m.ViberKeyboard)
	requestBody, response, err := viberBotApi.SendMessage(textMessage)
	if err != nil {
		err = errors.Wrap(err, "Failed to send message to Viber")
		logger.Errorf(c, err.Error())
	}
	log.Debugf(c, "Request body: %v", (string)(requestBody))
	if response.Status == 0 {
		logger.Debugf(c, "Succesfully sent to Viber")
	} else {
		switch response.Status { // https://developers.viber.com/customer/en/portal/articles/2541337-error-codes?b_id=15145
		case 2:
			logger.Errorf(c, "Viber response.Status=%v: %v: [%v]", response.Status, response.StatusMessage, botSettings.Token)
		default:
			logger.Errorf(c, "Viber response.Status=%v: %v", response.Status, response.StatusMessage)
		}
	}

	return resp, err
}
