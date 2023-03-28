package telegram

import (
	"github.com/strongo/dalgo/dal"
	"github.com/strongo/dalgo/record"
	"google.golang.org/appengine/datastore" // TODO: remove references to datastore
)

type Chat struct {
	record.WithID[string]
	//TgChatBase
	*ChatEntity
}

//var _ dal.EntityHolder = (*Chat)(nil)

func NewChat(id string) Chat {
	key := dal.NewKeyWithID(ChatKind, id)
	dto := new(ChatEntity)
	return Chat{
		WithID: record.WithID[string]{
			ID:     id,
			Record: dal.NewRecordWithData(key, dto),
		},
		ChatEntity: dto,
	}
}

func (Chat) Kind() string {
	return ChatKind
}

//func (tgChat Chat) Entity() interface{} {
//	return tgChat.ChatEntity
//}

//func (Chat) NewEntity() interface{} {
//	return new(ChatEntity)
//}

//func (tgChat *Chat) SetEntity(Data interface{}) {
//	if Data == nil {
//		tgChat.ChatEntity = nil
//	} else {
//		tgChat.ChatEntity = Data.(*ChatEntity)
//	}
//}

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
	//if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
	//	"TgChatInstanceID": gaedb.IsEmptyString,
	//}); err != nil {
	//	return
	//}
	return
}
