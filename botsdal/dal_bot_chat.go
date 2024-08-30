package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
)

const botChatsCollection = "botChats"

func NewBotChatKey(platformID, botID, chatID string) *dal.Key {
	botKey := NewBotKey(platformID, botID)
	return dal.NewKeyWithParentAndID(botKey, botChatsCollection, chatID)
}

// GetBotChat returns bot chat
// Deprecated: use
func GetBotChat(
	ctx context.Context,
	tx dal.ReadSession,
	platformID, botID, chatID string,
	newData func() botsfwmodels.BotChatData,
) (chat record.DataWithID[string, botsfwmodels.BotChatData], err error) {
	key := NewBotChatKey(platformID, botID, chatID)
	data := newData()
	chat = record.NewDataWithID(chatID, key, data)
	return chat, tx.Get(ctx, chat.Record)
}
