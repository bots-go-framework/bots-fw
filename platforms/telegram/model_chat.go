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

type TelegramChatEntity struct {
	bots.BotChatEntity
	TelegramUserID        int   `datastore:",noindex"`
	TelegramUserIDs       []int `datastore:",noindex"` // For groups
	LastProcessedUpdateID int   `datastore:",noindex"`
}

var _ bots.BotChat = (*TelegramChatEntity)(nil)

func NewTelegramChatEntity() *TelegramChatEntity {
	return &TelegramChatEntity{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: user.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}

func (entity *TelegramChatEntity) SetAppUserIntID(id int64) {
	entity.AppUserIntID = id
}

func (entity *TelegramChatEntity) SetBotUserID(id interface{}) {
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

func (entity *TelegramChatEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *TelegramChatEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.CleanProperties(properties); err != nil {
		return
	}
	return
}

func (_ *TelegramChatEntity) CleanProperties(properties []datastore.Property) ([]datastore.Property, error) {
	var err error
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
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
