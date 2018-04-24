package gae_host

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"context"
	"google.golang.org/appengine/datastore"
)

type tgChatInstanceDalGae struct {
}

var _ telegram_bot.TgChatInstanceDal = (*tgChatInstanceDalGae)(nil)

func NewTgChatInstanceKey(c context.Context, id string) *datastore.Key {
	return datastore.NewKey(c, telegram_bot.TelegramChatInstanceKind, id, 0, nil)
}
func (tgChatInstanceDalGae tgChatInstanceDalGae) GetTelegramChatInstanceByID(c context.Context, id string) (tgChatInstance telegram_bot.TelegramChatInstance, err error) {
	tgChatInstance = tgChatInstanceDalGae.NewTelegramChatInstance(id, 0, "")
	if err = gaedb.Get(c, NewTgChatInstanceKey(c, id), tgChatInstance.Entity()); err == datastore.ErrNoSuchEntity {
		tgChatInstance.SetEntity(nil)
		err = db.NewErrNotFoundByStrID(telegram_bot.TelegramChatInstanceKind, id, err)
		return
	}
	return
}

func (_ tgChatInstanceDalGae) SaveTelegramChatInstance(c context.Context, tgChatInstance telegram_bot.TelegramChatInstance) (err error) {
	if _, err = gaedb.Put(c, NewTgChatInstanceKey(c, tgChatInstance.ID), tgChatInstance.TelegramChatInstanceEntity); err != nil {
		err = errors.WithMessage(err, fmt.Sprintf(
			"failed to store to GAE datastore tgChatInstance.TelegramChatInstanceEntity: %T(%+v)",
			tgChatInstance.TelegramChatInstanceEntity, tgChatInstance.TelegramChatInstanceEntity))
	}
	return
}

func (_ tgChatInstanceDalGae) NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance telegram_bot.TelegramChatInstance) {
	return telegram_bot.TelegramChatInstance{
		StringID: db.StringID{ID: chatInstanceID},
		TelegramChatInstanceEntity: &TelegramChatInstanceEntityGae{
			TelegramChatInstanceEntityBase: telegram_bot.TelegramChatInstanceEntityBase{
				TgChatID:          chatID,
				PreferredLanguage: preferredLanguage,
			},
		},
	}
}

func init() {
	telegram_bot.DAL.TgChatInstance = tgChatInstanceDalGae{}
}
