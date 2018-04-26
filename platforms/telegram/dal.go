package telegram

import (
	"context"
	"github.com/strongo/db"
)

// TgChatInstanceDal is DAL for telegram chat instance entity
type TgChatInstanceDal interface {
	GetTelegramChatInstanceByID(c context.Context, id string) (tgChatInstance ChatInstance, err error)
	NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance ChatInstance)
	SaveTelegramChatInstance(c context.Context, tgChatInstance ChatInstance) (err error)
}

type dal struct {
	DB             db.Database
	TgChatInstance TgChatInstanceDal
}

// DAL is data access layer
var DAL dal
