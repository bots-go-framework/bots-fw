package viber_bot

import (
	"github.com/strongo/bots-api-viber/viberinterface"
	"github.com/strongo/bots-framework/core"
	"strconv"
	"time"
)

type viberWebhookMessage struct {
	ViberWebhookInput
	m    viberinterface.CallbackOnMessage // Can be either input.update.Message or input.update.CallbackQuery.Message
	chat ViberWebhookChat
}

func (whm viberWebhookMessage) IntID() int64 {
	return whm.m.MessageToken
}

func (whm viberWebhookMessage) StringID() string {
	return strconv.FormatInt(whm.m.MessageToken, 10)
}

func (whm viberWebhookMessage) BotChatID() (string, error) {
	return whm.chat.GetID(), nil
}

func (whm viberWebhookMessage) Chat() bots.WebhookChat {
	return whm.chat
}

func (whm viberWebhookMessage) GetRecipient() bots.WebhookRecipient {
	panic("Not supported (yet?)")
}

func (whm viberWebhookMessage) GetSender() bots.WebhookSender {
	return newViberSender(whm.m.Sender)
}

func (whm viberWebhookMessage) GetTime() time.Time {
	return time.Unix(whm.m.Timestamp, 0)
}

func newViberWebhookMessage(m viberinterface.CallbackOnMessage) viberWebhookMessage {
	return viberWebhookMessage{ViberWebhookInput: newViberWebhookInput(m.CallbackBase), m: m, chat: NewViberWebhookChat(m.Sender.ID)}
}
