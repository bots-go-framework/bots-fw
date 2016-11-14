package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"time"
)


type ViberWebhookInput struct {
	callbackBase viberinterface.CallbackBase
}

func newViberWebhookInput(callbackBase viberinterface.CallbackBase) ViberWebhookInput {
	return ViberWebhookInput{callbackBase: callbackBase}
}

func (whi ViberWebhookInput) GetTime() time.Time {
	return time.Unix(whi.callbackBase.Timestamp, 0)
}
