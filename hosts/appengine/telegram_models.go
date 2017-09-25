package gae_host

import (
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
)

type TelegramChatInstanceEntityGae struct {
	telegram_bot.TelegramChatInstanceEntityBase
}

var _ telegram_bot.TelegramChatInstanceEntity = (*TelegramChatInstanceEntityGae)(nil)

func (entity *TelegramChatInstanceEntityGae) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *TelegramChatInstanceEntityGae) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"PreferredLanguage":     gaedb.IsEmptyString,
	}); err != nil {
		return
	}
	return
}

func init() {
	telegram_bot.NewTelegramChatInstanceEntity = func() telegram_bot.TelegramChatInstanceEntity {
		return new(TelegramChatInstanceEntityGae)
	}
}