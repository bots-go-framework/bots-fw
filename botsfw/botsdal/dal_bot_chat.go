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
	return dal.NewKeyWithID(botChatsCollection, chatID, dal.WithParentKey(botKey))
}

func GetBotChat(
	ctx context.Context,
	tx dal.ReadSession,
	platformID, botID, chatID string,
	newData func() botsfwmodels.ChatData,
) (chat record.DataWithID[string, botsfwmodels.ChatData], err error) {
	key := NewBotChatKey(platformID, botID, chatID)
	data := newData()
	chat = record.NewDataWithID(chatID, key, data)
	return chat, tx.Get(ctx, chat.Record)
}
