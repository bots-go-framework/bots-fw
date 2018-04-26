package viber

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"time"
)

// webhookInput wrapper for Viber message
type webhookInput struct {
	callbackBase viberinterface.CallbackBase
}

// LogRequest logs request (not implemented yet)
func (webhookInput) LogRequest() {
	panic("Not implemented")
}

func newViberWebhookInput(callbackBase viberinterface.CallbackBase) webhookInput {
	return webhookInput{callbackBase: callbackBase}
}

// GetTime returns sent time of the message
func (whi webhookInput) GetTime() time.Time {
	return time.Unix(whi.callbackBase.Timestamp, 0)
}
