package telegram

import (
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type Chat struct {
	TgChatBase
	*ChatEntity
}

var _ db.EntityHolder = (*Chat)(nil)

func (Chat) Kind() string {
	return ChatKind
}

func (tgChat Chat) Entity() interface{} {
	return tgChat.ChatEntity
}

func (Chat) NewEntity() interface{} {
	return new(ChatEntity)
}

func (tgChat *Chat) SetEntity(entity interface{}) {
	if entity == nil {
		tgChat.ChatEntity = nil
	} else {
		tgChat.ChatEntity = entity.(*ChatEntity)
	}
}

type ChatEntity struct {
	UserGroupID string `datastore:",index,omitempty"` // Do index
	TgChatEntityBase
}

func (entity *ChatEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *ChatEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}
	if properties, err = entity.TgChatEntityBase.CleanProperties(properties); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"TgChatInstanceID": gaedb.IsEmptyString,
	}); err != nil {
		return
	}
	return
}

