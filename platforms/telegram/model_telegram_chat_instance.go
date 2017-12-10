package telegram_bot

import (
	"github.com/strongo/db"
)

const TelegramChatInstanceKind = "TgChatInstance"

type TelegramChatInstanceEntityBase struct {
	TgChatID          int64  `datastore:",noindex"`
	PreferredLanguage string `datastore:",noindex"`
}

type TelegramChatInstance struct {
	db.StringID
	TelegramChatInstanceEntity
}

var _ db.EntityHolder = (*TelegramChatInstance)(nil)

func (TelegramChatInstance) Kind() string {
	return TelegramChatInstanceKind
}

func (TelegramChatInstance) NewEntity() interface{} {
	panic("not supported")
}

func (record *TelegramChatInstance) Entity() interface{} {
	return record.TelegramChatInstanceEntity
}

var NewTelegramChatInstanceEntity func() TelegramChatInstanceEntity

func (record *TelegramChatInstance) SetEntity(entity interface{}) {
	if entity == nil {
		record.TelegramChatInstanceEntity = nil
	} else {
		record.TelegramChatInstanceEntity = entity.(TelegramChatInstanceEntity)
	}
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
