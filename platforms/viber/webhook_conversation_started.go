package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

type ViberWebhookInputConversationStarted struct {
	ViberWebhookInput
	m viberinterface.CallbackOnConversationStarted
	chat ViberWebhookChat
}

func (whi ViberWebhookInputConversationStarted) GetSender() bots.WebhookSender {
	return ViberSender{sender: whi.m.User.CallbackSender} // TODO: Extend to support User
}

func (whi ViberWebhookInputConversationStarted) GetContext() string {
	return whi.m.Context
}


func (whi ViberWebhookInputConversationStarted) GetRecipient() bots.WebhookRecipient {
	panic("GetRecipient() is not implemented yet or can not be supported at all.")
}

func (whi ViberWebhookInputConversationStarted) InputType() bots.WebhookInputType {
	return bots.WebhookInputConversationStarted
}

func (whi ViberWebhookInputConversationStarted) BotChatID(c context.Context) (chatID string, err error) {
	return whi.chat.GetID(), nil
}

func (whi ViberWebhookInputConversationStarted) Chat() bots.WebhookChat {
	return whi.chat
}

func NewViberWebhookInputConversationStarted(m viberinterface.CallbackOnConversationStarted) bots.WebhookInput {
	return ViberWebhookInputConversationStarted{
		m: m,
		chat: NewViberWebhookChat(m.User.ID),
		ViberWebhookInput: ViberWebhookInput{callbackBase: m.CallbackBase},
	}
}