package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
)

type callbackCurrent struct {
}

var CallbackCurrent bots.MessageUID = &callbackCurrent{}

func (_ callbackCurrent) UID() string {
	return "callbackCurrent"
}

type InlineMessageUID struct {
	InlineMessageID string
}

var _ bots.MessageUID = (*InlineMessageUID)(nil)

func NewInlineMessageUID(inlineMessageID string) *InlineMessageUID {
	return &InlineMessageUID{InlineMessageID: inlineMessageID}
}

func (m InlineMessageUID) UID() string {
	return m.InlineMessageID
}


func NewChatMessageUID(chatID int64, messageID int) *ChatMessageUID {
	return &ChatMessageUID{ChatID: chatID, MessageID: messageID}
}

type ChatMessageUID struct {
	ChatID int64
	MessageID int
}

var _ bots.MessageUID = (*ChatMessageUID)(nil)

func (m ChatMessageUID) UID() string {
	return fmt.Sprintf("%d:%d", m.ChatID, m.MessageID)
}
