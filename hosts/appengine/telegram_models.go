package gaehost

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
)

// TelegramChatInstanceEntityGae DAL to Telegram chat entity
type TelegramChatInstanceEntityGae struct {
	telegram.ChatInstanceEntityBase
}

var _ telegram.ChatInstanceEntity = (*TelegramChatInstanceEntityGae)(nil)

// Load Telegram chat entity
func (entity *TelegramChatInstanceEntityGae) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

// Save saves Telegram chat entity
func (entity *TelegramChatInstanceEntityGae) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}
	//if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
	//	"PreferredLanguage": gaedb.IsEmptyString,
	//}); err != nil {
	//	return
	//}
	return
}

func init() {
	telegram.NewChatInstanceEntity = func() telegram.ChatInstanceEntity {
		return new(TelegramChatInstanceEntityGae)
	}
}
