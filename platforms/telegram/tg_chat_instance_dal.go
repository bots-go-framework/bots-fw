package telegram

import (
	"context"
	"github.com/strongo/dalgo/dal"
	"github.com/strongo/db"
)

type tgChatInstanceDalgo struct {
	db dal.Database
}

var _ TgChatInstanceDal = (*tgChatInstanceDalgo)(nil)

func (tgChatInstanceDal tgChatInstanceDalgo) GetTelegramChatInstanceByID(c context.Context, tx dal.ReadTransaction, id string) (tgChatInstance ChatInstance, err error) {
	tgChatInstance = tgChatInstanceDal.NewTelegramChatInstance(id, 0, "")

	var session dal.ReadSession
	if tx == nil {
		session = tgChatInstanceDal.db
	} else {
		session = tx
	}
	if err = session.Get(c, tgChatInstance.record); dal.IsNotFound(err) {
		tgChatInstance.SetEntity(nil)
		return
	}
	return
}

func (tgChatInstanceDal tgChatInstanceDalgo) SaveTelegramChatInstance(c context.Context, tgChatInstance ChatInstance) (err error) {
	err = tgChatInstanceDal.db.RunReadwriteTransaction(c, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, tgChatInstance.record)
	})
	return
}

func (tgChatInstanceDalgo) NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance ChatInstance) {
	tgChatInstance = ChatInstance{
		StringID: db.StringID{ID: chatInstanceID},
	}
	tgChatInstance.SetEntity(&ChatInstanceEntityBase{
		TgChatID:          chatID,
		PreferredLanguage: preferredLanguage,
	})
	return tgChatInstance
}

func init() {
	DAL.TgChatInstance = tgChatInstanceDalgo{}
}
