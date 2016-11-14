package fbm_strongo_bot

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-fbm"
	"time"
	"fmt"
)

type FbmWebhookInput struct {
	messaging fbm_bot_api.Messaging
}

var _ bots.WebhookInput = (*FbmWebhookInput)(nil)

func (whi FbmWebhookInput) Chat() bots.WebhookChat {
	return FbmWebhookChat{
		ID: fmt.Sprintf("%v-%v", whi.messaging.Sender.ID, whi.messaging.Recipient.ID),
	}
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
	return whi.messaging.Message
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
