package telegram

import (
	"github.com/strongo/dalgo/dal"
	"github.com/strongo/dalgo/record"
)

// ChatInstanceKind is kind name of TgChatInstance Data
const ChatInstanceKind = "TgChatInstance"

// ChatInstanceEntityBase is base struct
type ChatInstanceEntityBase struct {
	TgChatID          int64  `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
}

func NewTgChatInstanceKey(id string) *dal.Key {
	return dal.NewKey(ChatInstanceKind, dal.WithStringID(id))
}

// ChatInstance is base struct
type ChatInstance struct {
	record.WithID[string]
	dal.Record
	Data ChatInstanceEntity
}

//var _ db.EntityHolder = (*ChatInstance)(nil)

//var _ dal.Record = (*ChatInstance)(nil)

//// NewEntity creates new Data
//func (ChatInstance) NewEntity() interface{} {
//	panic("not supported")
//}
//
//// Entity returns Data for saving
//func (record *ChatInstance) Entity() interface{} {
//	return record.Data
//}

// NewChatInstanceEntity is pointer to func() ChatInstanceEntity
var NewChatInstanceEntity func() ChatInstanceEntity

// SetEntity sets Data to record
func (record *ChatInstance) SetEntity(entity interface{}) {
	record.Data = entity.(ChatInstanceEntity)
	//if Data == nil {
	//	record.Entity = nil
	//} else {
	//	record.Entity = Data.(ChatInstanceEntity)
	//}
}

//func (record *ChatInstance) SetStrID(id string) {
//	record.ID = id
//}

// ChatInstanceEntity describes chat instance Data interface
type ChatInstanceEntity interface {
	GetTgChatID() int64
	GetPreferredLanguage() string
	SetPreferredLanguage(v string)
}

// GetTgChatID returns Telegram chat ID
func (entity ChatInstanceEntityBase) GetTgChatID() int64 {
	return entity.TgChatID
}

// GetPreferredLanguage returns preferred language for the chat
func (entity ChatInstanceEntityBase) GetPreferredLanguage() string {
	return entity.PreferredLanguage
}

// SetPreferredLanguage sets preferred language for the chat
func (entity *ChatInstanceEntityBase) SetPreferredLanguage(v string) {
	entity.PreferredLanguage = v
}
