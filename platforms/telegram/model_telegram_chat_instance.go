package telegram_bot

import (
	"github.com/strongo/app/db"
)

const TelegramChatInstanceKind = "TgChatInstance"

type TelegramChatInstanceEntityBase struct {
	TgChatID int64 `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
}

type TelegramChatInstance struct {
	ID string
	db.NoIntID
	TelegramChatInstanceEntity
}

func (TelegramChatInstance) Kind() string {
	return TelegramChatInstanceKind
}

func (record TelegramChatInstance) StrID() string {
	return record.ID
}

func (record *TelegramChatInstance) SetStrID(id string) {
	record.ID = id
}

var _ db.EntityHolder = (*TelegramChatInstance)(nil)

func (record *TelegramChatInstance) Entity() interface{} {
	if record.TelegramChatInstanceEntity == nil {
		if NewTelegramChatInstanceEntity == nil {
			panic("NewTelegramChatInstanceEntity is nil")
		}
		record.TelegramChatInstanceEntity = NewTelegramChatInstanceEntity()
	}
	return record.TelegramChatInstanceEntity
}

var NewTelegramChatInstanceEntity func() TelegramChatInstanceEntity

func (record *TelegramChatInstance) SetEntity(entity interface{}) {
	record.TelegramChatInstanceEntity = entity.(TelegramChatInstanceEntity)
}

//func (record *TelegramChatInstance) SetStrID(id string) {
//	record.ID = id
//}


type TelegramChatInstanceEntity interface {
	GetTgChatID() int64
	GetPreferredLanguage() string
	SetPreferredLanguage(v string)
}

func (entity TelegramChatInstanceEntityBase) GetTgChatID() int64 {
	return entity.TgChatID
}

func (entity TelegramChatInstanceEntityBase) GetPreferredLanguage() string {
	return entity.PreferredLanguage
}

func (entity *TelegramChatInstanceEntityBase) SetPreferredLanguage(v string) {
	entity.PreferredLanguage = v
}
