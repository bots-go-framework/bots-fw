package telegram

import (
	"github.com/strongo/bots-framework/core"
)

type tgWebhookNewChatMembersMessage struct {
	tgWebhookMessage
}

func (tgWebhookNewChatMembersMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputNewChatMembers
}

var _ bots.WebhookNewChatMembersMessage = (*tgWebhookNewChatMembersMessage)(nil)

func newTgWebhookNewChatMembersMessage(input tgWebhookInput) tgWebhookNewChatMembersMessage {
	return tgWebhookNewChatMembersMessage{tgWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m tgWebhookNewChatMembersMessage) NewChatMembers() []bots.WebhookActor {
	members := make([]bots.WebhookActor, len(m.message.NewChatMembers))
	for i, m := range m.message.NewChatMembers {
		members[i] = m
	}
	return members
}
