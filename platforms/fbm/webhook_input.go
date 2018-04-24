package fbm_bot

import (
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"time"
)

// FbmWebhookInput provides information on current FBM message
type FbmWebhookInput struct {
	messaging fbm_api.Messaging
}

var _ bots.WebhookInput = (*FbmWebhookInput)(nil)
var _ bots.WebhookMessage = (*FbmWebhookInput)(nil)

// IntID is not supported, use StringID
func (_ FbmWebhookInput) IntID() int64 {
	panic("Not supported")
}

// LogRequest is not implemented yet
func (_ FbmWebhookInput) LogRequest() {
	panic("Not implemented")
}

// StringID returns an unique FBM message ID
func (whi FbmWebhookInput) StringID() string {
	return whi.messaging.Message.MID
}

// BotChatID returns FBM chat ID for the message
func (whi FbmWebhookInput) BotChatID() (string, error) {
	return whi.messaging.Sender.ID, nil
}

// BotChat returns instance of FBM chat from app storage
func (whi FbmWebhookInput) Chat() bots.WebhookChat {
	return NewFbmWebhookChat(whi.messaging.Sender.ID)
}

// GetSender returns information about sender of the FBM message
func (whi FbmWebhookInput) GetSender() bots.WebhookSender {
	return whi.messaging.Sender
}

// GetRecipient returns information about receiver of the FBM message
func (whi FbmWebhookInput) GetRecipient() bots.WebhookRecipient {
	return whi.messaging.Recipient
}

// GetTime returns when the mesage was sent
func (whi FbmWebhookInput) GetTime() time.Time {
	return time.Unix(whi.messaging.Timestamp, 0)
}

// GetTime returns the input message
func (whi FbmWebhookInput) InputMessage() bots.WebhookMessage {
	panic("Not implemented return whi.messaging.Message") // TODO: Do we really need .Chat() in Message interface?
}

// InputPostback is not supported or not implemented
func (whi FbmWebhookInput) InputPostback() bots.WebhookPostback {
	return nil
}

// InputDelivery is not supported or not implemented
func (whi FbmWebhookInput) InputDelivery() bots.WebhookDelivery {
	return nil
}

// InputInlineQuery is not supported
func (whi FbmWebhookInput) InputInlineQuery() bots.WebhookInlineQuery {
	panic("Not supported")
}

// InputCallbackQuery is not implemented yet
func (whi FbmWebhookInput) InputCallbackQuery() bots.WebhookCallbackQuery {
	panic("Not implemented")
}

// InputChosenInlineResult is not supported
func (whi FbmWebhookInput) InputChosenInlineResult() bots.WebhookChosenInlineResult {
	panic("Not supported")
}

// InputType returns type of the message
func (whi FbmWebhookInput) InputType() bots.WebhookInputType {
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

// FbmTextMessage provides information about text message
type FbmTextMessage struct {
	FbmWebhookInput
}

// Text returns text of the message
func (textMessage FbmTextMessage) Text() string {
	return textMessage.messaging.Message.Text()
}

var _ bots.WebhookTextMessage = (*FbmTextMessage)(nil)

// NewFbmWebhookInput maps API struct to framework struct
func NewFbmWebhookInput(messaging fbm_api.Messaging) bots.WebhookInput {
	fbmInput := FbmWebhookInput{messaging: messaging}
	switch {
	case messaging.Message != nil:
		return FbmTextMessage{FbmWebhookInput: fbmInput}
	case messaging.Postback != nil:
		return FbmPostbackInput{FbmWebhookInput: fbmInput}
	}
	return fbmInput
}

// IsEdited indicates if message was edited. Always false for FBM
func (whm FbmTextMessage) IsEdited() bool {
	return false
}
