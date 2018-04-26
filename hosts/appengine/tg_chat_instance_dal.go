package gaehost

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type tgChatInstanceDalGae struct {
}

var _ telegram.TgChatInstanceDal = (*tgChatInstanceDalGae)(nil)

func newTgChatInstanceKey(c context.Context, id string) *datastore.Key {
	return datastore.NewKey(c, telegram.ChatInstanceKind, id, 0, nil)
}
func (tgChatInstanceDalGae tgChatInstanceDalGae) GetTelegramChatInstanceByID(c context.Context, id string) (tgChatInstance telegram.ChatInstance, err error) {
	tgChatInstance = tgChatInstanceDalGae.NewTelegramChatInstance(id, 0, "")
	if err = gaedb.Get(c, newTgChatInstanceKey(c, id), tgChatInstance.Entity()); err == datastore.ErrNoSuchEntity {
		tgChatInstance.SetEntity(nil)
		err = db.NewErrNotFoundByStrID(telegram.ChatInstanceKind, id, err)
		return
	}
	return
}

func (tgChatInstanceDalGae) SaveTelegramChatInstance(c context.Context, tgChatInstance telegram.ChatInstance) (err error) {
	if _, err = gaedb.Put(c, newTgChatInstanceKey(c, tgChatInstance.ID), tgChatInstance.ChatInstanceEntity); err != nil {
		err = errors.WithMessage(err, fmt.Sprintf(
			"failed to store to GAE datastore tgChatInstance.ChatInstanceEntity: %T(%+v)",
			tgChatInstance.ChatInstanceEntity, tgChatInstance.ChatInstanceEntity))
	}
	return
}

func (tgChatInstanceDalGae) NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance telegram.ChatInstance) {
	return telegram.ChatInstance{
		StringID: db.StringID{ID: chatInstanceID},
		ChatInstanceEntity: &TelegramChatInstanceEntityGae{
			ChatInstanceEntityBase: telegram.ChatInstanceEntityBase{
				TgChatID:          chatID,
				PreferredLanguage: preferredLanguage,
			},
		},
	}
}

func init() {
	telegram.DAL.TgChatInstance = tgChatInstanceDalGae{}
}
