package fbm_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-fbm"
	"time"
)

type FbmWebhookInput struct {
	messaging fbm_bot_api.Messaging
}

var _ bots.WebhookInput = (*FbmWebhookInput)(nil)
var _ bots.WebhookMessage = (*FbmWebhookInput)(nil)

func (_ FbmWebhookInput) IntID() int64 {
	panic("Not supported")
}

func (self FbmWebhookInput) StringID() string {
	return self.messaging.Message.MID
}


func (whi FbmWebhookInput) Chat() bots.WebhookChat {
	return NewFbmWebhookChat(whi.messaging.Sender.ID)
}

func (whi FbmWebhookInput) GetSender() bots.WebhookSender {
	return whi.messaging.Sender
}

func (whi FbmWebhookInput) GetRecipient() bots.WebhookRecipient {
	return whi.messaging.Recipient
}

func (whi FbmWebhookInput) GetTime() time.Time {
	return time.Unix(whi.messaging.Timestamp, 0)
}

func (whi FbmWebhookInput) InputMessage() bots.WebhookMessage {
	panic("Not implemented return whi.messaging.Message") // TODO: Do we really need .Chat() in Message interface?
}

func (whi FbmWebhookInput) InputPostback() bots.WebhookPostback {
	return nil
}

func (whi FbmWebhookInput) InputDelivery() bots.WebhookDelivery {
	return nil
}

func (whi FbmWebhookInput) InputInlineQuery() bots.WebhookInlineQuery {
	panic("Not implemented")
}

func (whi FbmWebhookInput) InputCallbackQuery() bots.WebhookCallbackQuery {
	panic("Not implemented")
}

func (whi FbmWebhookInput) InputChosenInlineResult() bots.WebhookChosenInlineResult {
	panic("Not implemented")
}

func (whi FbmWebhookInput) InputType() bots.WebhookInputType {
	switch {
	case whi.messaging.Message != nil:
		if len(whi.messaging.Message.Attachments) > 0 {
			return bots.WebhookInputAttachment
		} else if len(whi.messaging.Message.MText) > 0 {
			return bots.WebhookInputText
		}
	case whi.messaging.Postback != nil:
		return bots.WebhookInputPostback
	case whi.messaging.Delivery != nil:
		return bots.WebhookInputDelivery
	}
	return bots.WebhookInputUnknown
}

type FbmTextMessage struct {
	FbmWebhookInput
}

func (textMessage FbmTextMessage) Text() string {
	return textMessage.messaging.Message.Text()
}

var _ bots.WebhookTextMessage = (*FbmTextMessage)(nil)


func NewFbmWebhookInput(messaging fbm_bot_api.Messaging) bots.WebhookInput {
	fbmInput := FbmWebhookInput{messaging: messaging}
	return FbmTextMessage{FbmWebhookInput: fbmInput}
}