package fbm

import (
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"time"
)

// webhookInput provides information on current FBM message
type webhookInput struct {
	messaging fbmbotapi.Messaging
}

var _ bots.WebhookInput = (*webhookInput)(nil)
var _ bots.WebhookMessage = (*webhookInput)(nil)

// IntID is not supported, use StringID
func (webhookInput) IntID() int64 {
	panic("Not supported")
}

// LogRequest is not implemented yet
func (webhookInput) LogRequest() {
	panic("Not implemented")
}

// StringID returns an unique FBM message ID
func (whi webhookInput) StringID() string {
	return whi.messaging.Message.MID
}

// BotChatID returns FBM chat ID for the message
func (whi webhookInput) BotChatID() (string, error) {
	return whi.messaging.Sender.ID, nil
}

// Chat returns instance of FBM chat from app storage
func (whi webhookInput) Chat() bots.WebhookChat {
	return NewFbmWebhookChat(whi.messaging.Sender.ID)
}

// GetSender returns information about sender of the FBM message
func (whi webhookInput) GetSender() bots.WebhookSender {
	return whi.messaging.Sender
}

// GetRecipient returns information about receiver of the FBM message
func (whi webhookInput) GetRecipient() bots.WebhookRecipient {
	return whi.messaging.Recipient
}

// GetTime returns when the mesage was sent
func (whi webhookInput) GetTime() time.Time {
	return time.Unix(whi.messaging.Timestamp, 0)
}

// InputMessage returns the input message
func (whi webhookInput) InputMessage() bots.WebhookMessage {
	panic("Not implemented return whi.messaging.Message") // TODO: Do we really need .Chat() in Message interface?
}

// InputPostback is not supported or not implemented
func (whi webhookInput) InputPostback() bots.WebhookPostback {
	return nil
}

// InputDelivery is not supported or not implemented
func (whi webhookInput) InputDelivery() bots.WebhookDelivery {
	return nil
}

// InputInlineQuery is not supported
func (whi webhookInput) InputInlineQuery() bots.WebhookInlineQuery {
	panic("Not supported")
}

// InputCallbackQuery is not implemented yet
func (whi webhookInput) InputCallbackQuery() bots.WebhookCallbackQuery {
	panic("Not implemented")
}

// InputChosenInlineResult is not supported
func (whi webhookInput) InputChosenInlineResult() bots.WebhookChosenInlineResult {
	panic("Not supported")
}

// InputType returns type of the message
func (whi webhookInput) InputType() bots.WebhookInputType {
	switch {
	case whi.messaging.Message != nil:
		if len(whi.messaging.Message.Attachments) > 0 {
			return bots.WebhookInputAttachment
		} else if len(whi.messaging.Message.MText) > 0 {
			return bots.WebhookInputText
		}
	case whi.messaging.Postback != nil:
		return bots.WebhookInputCallbackQuery
	case whi.messaging.Delivery != nil:
		return bots.WebhookInputDelivery
	}
	return bots.WebhookInputUnknown
}

// textMessage provides information about text message
type textMessage struct {
	webhookInput
}

// Text returns text of the message
func (textMessage textMessage) Text() string {
	return textMessage.messaging.Message.Text()
}

var _ bots.WebhookTextMessage = (*textMessage)(nil)

// NewFbmWebhookInput maps API struct to framework struct
func NewFbmWebhookInput(messaging fbmbotapi.Messaging) bots.WebhookInput {
	fbmInput := webhookInput{messaging: messaging}
	switch {
	case messaging.Message != nil:
		return textMessage{webhookInput: fbmInput}
	case messaging.Postback != nil:
		return postbackInput{webhookInput: fbmInput}
	}
	return fbmInput
}

// IsEdited indicates if message was edited. Always false for FBM
func (textMessage textMessage) IsEdited() bool {
	return false
}
