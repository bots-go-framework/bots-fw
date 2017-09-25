package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"time"
	"strconv"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/user"
)

const (
	TelegramChatKind = "TgChat"
)

type TelegramChatEntity interface {
	SetTgChatInstanceID(v string)
	GetTgChatInstanceID() string
	GetPreferredLanguage() string
}

type TelegramChatEntityBase struct {
	bots.BotChatEntity
	TelegramUserID        int    `datastore:",noindex"`
	TelegramUserIDs       []int  `datastore:",noindex"` // For groups
	LastProcessedUpdateID int    `datastore:",noindex"`
	TgChatInstanceID      string `datastore:",noindex"` // Do index
}

func (entity *TelegramChatEntityBase) SetTgChatInstanceID(v string) {
	entity.TgChatInstanceID = v
}

func (entity *TelegramChatEntityBase) GetTgChatInstanceID() string {
	return entity.TgChatInstanceID
}

func (entity *TelegramChatEntityBase) GetPreferredLanguage() string {
	return entity.PreferredLanguage
}

var _ bots.BotChat = (*TelegramChatEntityBase)(nil)

func NewTelegramChatEntity() *TelegramChatEntityBase {
	return &TelegramChatEntityBase{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: user.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}

func (entity *TelegramChatEntityBase) SetAppUserIntID(id int64) {
	if entity.IsGroup && id != 0 {
		panic("TelegramChatEntityBase.IsGroup && id != 0")
	}
	entity.AppUserIntID = id
}

func (entity *TelegramChatEntityBase) SetBotUserID(id interface{}) {
	switch id.(type) {
	case string:
		var err error
		entity.TelegramUserID, err = strconv.Atoi(id.(string))
		if err != nil {
			panic(err.Error())
		}
	case int:
		entity.TelegramUserID = id.(int)
	case int64:
		entity.TelegramUserID = id.(int)
	default:
		panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
	}
}

func (entity *TelegramChatEntityBase) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *TelegramChatEntityBase) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.CleanProperties(properties); err != nil {
		return
	}
	return
}

func (entity *TelegramChatEntityBase) CleanProperties(properties []datastore.Property) ([]datastore.Property, error) {
	if entity.IsGroup && entity.AppUserIntID != 0 {
		panic(fmt.Sprintf("IsGroup && AppUserIntID:%d != 0", entity.AppUserIntID))
	}
	var err error
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"AppUserIntID":          gaedb.IsZeroInt,
		"AccessGranted":         gaedb.IsFalse,
		"AwaitingReplyTo":       gaedb.IsEmptyString,
		"DtForbidden":           gaedb.IsZeroTime,
		"DtForbiddenLast":       gaedb.IsZeroTime,
		"GaClientID":            gaedb.IsEmptyByteArray,
		"TelegramUserID":        gaedb.IsZeroInt,
		"LastProcessedUpdateID": gaedb.IsZeroInt,
		"PreferredLanguage":     gaedb.IsEmptyString,
		"Title":                 gaedb.IsEmptyString, // TODO: Is it obsolete?
		"Type":                  gaedb.IsEmptyString, // TODO: Is it obsolete?
	}); err != nil {
		return properties, err
	}
	return properties, err
}
