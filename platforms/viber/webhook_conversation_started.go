package viber

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
)

// WebhookInputConversationStarted is Viber message for new conversation
type WebhookInputConversationStarted struct { // TODO: make private
	webhookInput
	m    viberinterface.CallbackOnConversationStarted
	chat viberWebhookChat
}

// GetSender returns sender
func (whi WebhookInputConversationStarted) GetSender() bots.WebhookSender {
	return Sender{sender: whi.m.User.CallbackSender} // TODO: Extend to support User
}

// GetContext returns context of the message
func (whi WebhookInputConversationStarted) GetContext() string {
	return whi.m.Context
}

// GetRecipient return addressed receiver of the message
func (whi WebhookInputConversationStarted) GetRecipient() bots.WebhookRecipient {
	panic("GetRecipient() is not implemented yet or can not be supported at all.")
}

// InputType returns WebhookInputConversationStarted
func (whi WebhookInputConversationStarted) InputType() bots.WebhookInputType {
	return bots.WebhookInputConversationStarted
}

// BotChatID returns Viber chat ID
func (whi WebhookInputConversationStarted) BotChatID() (string, error) {
	return whi.chat.GetID(), nil
}

// Chat returns Viber chat wrapper
func (whi WebhookInputConversationStarted) Chat() bots.WebhookChat {
	return whi.chat
}

func newViberWebhookInputConversationStarted(m viberinterface.CallbackOnConversationStarted) bots.WebhookInput {
	return WebhookInputConversationStarted{
		m:            m,
		chat:         newViberWebhookChat(m.User.ID),
		webhookInput: webhookInput{callbackBase: m.CallbackBase},
	}
}
