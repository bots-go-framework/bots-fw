package fbm_bot

import "github.com/strongo/bots-framework/core"

type FbmPostbackInput struct {
	FbmWebhookInput
}

var _ bots.WebhookCallbackQuery = (*FbmPostbackInput)(nil)

func (input FbmPostbackInput) GetID() interface{} {
	return input.messaging.Timestamp
}

func (input FbmPostbackInput) GetInlineMessageID() string {
	return ""
}

func (input FbmPostbackInput) GetFrom() bots.WebhookSender {
	return input.FbmWebhookInput.GetSender()
}

func (input FbmPostbackInput) GetData() string {
	return input.messaging.Postback.Payload
}

func (input FbmPostbackInput) GetMessage() bots.WebhookMessage {
	return input.FbmWebhookInput
}
