package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
)

type TelegramWebhookLeftChatMembersMessage struct {
	telegramWebhookMessage
}

func (_ TelegramWebhookLeftChatMembersMessage) InputType() bots.WebhookInputType {
	return bots.WebhookInputLeftChatMembers
}

var _ bots.WebhookLeftChatMembersMessage = (*TelegramWebhookLeftChatMembersMessage)(nil)

func NewTelegramWebhookLeftChatMembersMessage(input telegramWebhookInput) TelegramWebhookNewChatMembersMessage {
	return TelegramWebhookNewChatMembersMessage{telegramWebhookMessage: newTelegramWebhookMessage(input, input.update.Message)}
}

func (m *TelegramWebhookLeftChatMembersMessage) LeftChatMembers() []bots.WebhookActor {
	return []bots.WebhookActor{m.message.LeftChatMember}
}
