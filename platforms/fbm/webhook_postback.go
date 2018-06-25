package fbm

import (
	"github.com/strongo/bots-framework/core"
	"strconv"
)

// postbackInput is postback input
type postbackInput struct {
	webhookInput
}

var _ bots.WebhookCallbackQuery = (*postbackInput)(nil)

// GetID returns ID
func (input postbackInput) GetID() string {
	return strconv.FormatInt(input.messaging.Timestamp, 10)
}

// GetInlineMessageID is not supported by FBM
func (input postbackInput) GetInlineMessageID() string {
	return ""
}

// GetFrom returns sender
func (input postbackInput) GetFrom() bots.WebhookSender {
	return input.webhookInput.GetSender()
}

// GetData returns payload
func (input postbackInput) GetData() string {
	return input.messaging.Postback.Payload
}

// GetMessage returns message
func (input postbackInput) GetMessage() bots.WebhookMessage {
	return input.webhookInput
}
