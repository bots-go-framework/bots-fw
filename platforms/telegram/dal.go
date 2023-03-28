package telegram

import (
	"context"
	"github.com/strongo/dalgo/dal"
)

// TgChatInstanceDal is DAL for telegram chat instance Data
type TgChatInstanceDal interface {
	GetTelegramChatInstanceByID(c context.Context, tx dal.ReadTransaction, id string) (tgChatInstance ChatInstance, err error)
	NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance ChatInstance)
	SaveTelegramChatInstance(c context.Context, tgChatInstance ChatInstance) (err error)
}

type dal1 struct {
	DB             dal.Database
	TgChatInstance TgChatInstanceDal
}

// DAL is data access layer
var DAL dal1
