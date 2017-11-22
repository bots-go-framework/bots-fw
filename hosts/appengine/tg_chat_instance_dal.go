package gae_host

import (
	"golang.org/x/net/context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/db"
	"github.com/strongo/bots-framework/platforms/telegram"
)

type tgChatInstanceDalGae struct {
}

var _ telegram_bot.TgChatInstanceDal = (*tgChatInstanceDalGae)(nil)

func NewTgChatInstanceKey(c context.Context, id string) *datastore.Key {
	return datastore.NewKey(c, telegram_bot.TelegramChatInstanceKind, id, 0, nil)
}
func (_ tgChatInstanceDalGae) GetTelegramChatInstanceByID(c context.Context, id string) (tgChatInstance telegram_bot.TelegramChatInstance, err error) {
	if err = gaedb.Get(c, NewTgChatInstanceKey(c, id), tgChatInstance.Entity()); err == datastore.ErrNoSuchEntity {
		err = db.NewErrNotFoundByStrID(telegram_bot.TelegramChatInstanceKind, id, err)
	}
	return
}

func (_ tgChatInstanceDalGae) SaveTelegramChatInstance(c context.Context, tgChatInstance telegram_bot.TelegramChatInstance) (err error) {
	_, err = gaedb.Put(c, NewTgChatInstanceKey(c, tgChatInstance.ID), tgChatInstance.TelegramChatInstanceEntity)
	return
}

func (_ tgChatInstanceDalGae) NewTelegramChatInstance(chatInstanceID string, chatID int64, preferredLanguage string) (tgChatInstance telegram_bot.TelegramChatInstance) {
	return telegram_bot.TelegramChatInstance{
		StringID: db.StringID{ID: chatInstanceID},
		TelegramChatInstanceEntity: &TelegramChatInstanceEntityGae{
			TelegramChatInstanceEntityBase: telegram_bot.TelegramChatInstanceEntityBase{
				 TgChatID: chatID,
				 PreferredLanguage: preferredLanguage,
			},
		},
	}
}

func init() {
	telegram_bot.DAL.TgChatInstance = tgChatInstanceDalGae{}
}