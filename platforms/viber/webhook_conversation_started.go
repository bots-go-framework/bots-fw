package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

type ViberWebhookInputConversationStarted struct {
	ViberWebhookInput
	m viberinterface.CallbackOnConversationStarted
}

func (whi ViberWebhookInputConversationStarted) GetSender() bots.WebhookSender {
	return ViberSender{sender: whi.m.User.CallbackSender} // TODO: Extend to support User
}

func (whi ViberWebhookInputConversationStarted) GetRecipient() bots.WebhookRecipient {
	panic("No")
}

func (whi ViberWebhookInputConversationStarted) InputType() bots.WebhookInputType {
	return bots.WebhookInputConversationStarted
}

func (whi ViberWebhookInputConversationStarted) Chat() bots.WebhookChat {
	panic("")
}

func NewViberWebhookInputConversationStarted(m viberinterface.CallbackOnConversationStarted) bots.WebhookInput {
	return ViberWebhookInputConversationStarted{
		m: m,
		ViberWebhookInput: ViberWebhookInput{callbackBase: m.CallbackBase},
	}
}