package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"time"
	"strconv"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/app/user"
	"github.com/strongo/db"
)

const (
	TelegramChatKind = "TgChat"
)

type TelegramChatEntity interface {
	SetTgChatInstanceID(v string)
	GetTgChatInstanceID() string
	GetPreferredLanguage() string
}

type TelegramChatBase struct {
	db.StringID
}

func (tgChat *TelegramChatBase) SetID(tgBotID string, tgChatID int64) {
	tgChat.ID = tgBotID + ":" + strconv.FormatInt(tgChatID, 10) // TODO: Should we migrated to format "id@bot"?
}

type TelegramChatEntityBase struct {
	bots.BotChatEntity
	TelegramUserID        int64    `datastore:",noindex"`
	TelegramUserIDs       []int64  `datastore:",noindex"` // For groups
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
		entity.TelegramUserID, err = strconv.ParseInt(id.(string), 10, 64)
		if err != nil {
			panic(err.Error())
		}
	case int:
		entity.TelegramUserID = int64(id.(int))
	case int64:
		entity.TelegramUserID = id.(int64)
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
		for _, userID := range entity.AppUserIntIDs {
			if userID == entity.AppUserIntID {
				goto found
			}
		}
		entity.AppUserIntIDs = append(entity.AppUserIntIDs, entity.AppUserIntID)
		entity.AppUserIntID = 0
	found:
	}

	for i, userID := range entity.AppUserIntIDs {
		if userID == 0 {
			panic(fmt.Sprintf("*TelegramChatEntityBase.AppUserIntIDs[%d] == 0", i))
		}
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
