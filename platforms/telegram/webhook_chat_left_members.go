package telegram

import (
	"github.com/strongo/bots-framework/core"
)

type tgWebhookLeftChatMembersMessage struct {
	tgWebhookMessage
}

func (tgWebhookLeftChatMembersMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputLeftChatMembers
}

var _ bots.WebhookLeftChatMembersMessage = (*tgWebhookLeftChatMembersMessage)(nil)

func newTgWebhookLeftChatMembersMessage(input tgWebhookInput) tgWebhookNewChatMembersMessage {
	return tgWebhookNewChatMembersMessage{tgWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m *tgWebhookLeftChatMembersMessage) LeftChatMembers() []bots.WebhookActor {
	return []bots.WebhookActor{m.message.LeftChatMember}
}
