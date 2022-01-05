package telegram

import "github.com/strongo/db"

// ChatInstanceKind is kind name of TgChatInstance entity
const ChatInstanceKind = "TgChatInstance"

// ChatInstanceEntityBase is base struct
type ChatInstanceEntityBase struct {
	TgChatID          int64  `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
}

// ChatInstance is base struct
type ChatInstance struct {
	db.StringID
	ChatInstanceEntity
}

var _ db.EntityHolder = (*ChatInstance)(nil)

// Kind return ChatInstanceKind
func (ChatInstance) Kind() string {
	return ChatInstanceKind
}

// NewEntity creates new entity
func (ChatInstance) NewEntity() interface{} {
	panic("not supported")
}

// Entity returns entity for saving
func (record *ChatInstance) Entity() interface{} {
	return record.ChatInstanceEntity
}

// NewChatInstanceEntity is pointer to func() ChatInstanceEntity
var NewChatInstanceEntity func() ChatInstanceEntity

// SetEntity sets entity to record
func (record *ChatInstance) SetEntity(entity interface{}) {
	if entity == nil {
		record.ChatInstanceEntity = nil
	} else {
		record.ChatInstanceEntity = entity.(ChatInstanceEntity)
	}
}

//func (record *ChatInstance) SetStrID(id string) {
//	record.ID = id
//}

// ChatInstanceEntity describes chat instance entity interface
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
