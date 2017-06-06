package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookNewChatMembersMessage struct {
	telegramWebhookMessage
}

func (_ TelegramWebhookNewChatMembersMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputNewChatMembers
}

var _ bots.WebhookNewChatMembersMessage = (*TelegramWebhookNewChatMembersMessage)(nil)

func NewTelegramWebhookNewChatMembersMessage(input TelegramWebhookInput) TelegramWebhookNewChatMembersMessage {
	return TelegramWebhookNewChatMembersMessage{telegramWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m TelegramWebhookNewChatMembersMessage) NewChatMembers() []bots.WebhookActor {
	members := make([]bots.WebhookActor, len(m.message.NewChatMembers))
	for i, m := range m.message.NewChatMembers {
		members[i] = m
	}
	return members
}
