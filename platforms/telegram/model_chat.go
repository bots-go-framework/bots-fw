package telegram

import (
	"fmt"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
	"strconv"
	"time"
)

const (
	// ChatKind is kind name of Telegram chat entity
	ChatKind = "TgChat"
)

// TgChatEntity is Telegram chat entity interface
type TgChatEntity interface {
	SetTgChatInstanceID(v string)
	GetTgChatInstanceID() string
	GetPreferredLanguage() string
}

// TgChatBase holds base properties of Telegram chat entity
type TgChatBase struct {
	db.StringID
}

// SetID sets ID
func (tgChat *TgChatBase) SetID(tgBotID string, tgChatID int64) {
	tgChat.ID = tgBotID + ":" + strconv.FormatInt(tgChatID, 10) // TODO: Should we migrated to format "id@bot"?
}

// TgChatEntityBase holds base properties of Telegram chat entity
type TgChatEntityBase struct {
	bots.BotChatEntity
	TelegramUserID        int64   `datastore:",noindex,omitempty"`
	TelegramUserIDs       []int64 `datastore:",noindex"` // For groups
	LastProcessedUpdateID int     `datastore:",noindex,omitempty"`
	TgChatInstanceID      string  `datastore:",noindex,omitempty"` // Do index
}

// SetTgChatInstanceID is what it is
func (entity *TgChatEntityBase) SetTgChatInstanceID(v string) {
	entity.TgChatInstanceID = v
}

// GetTgChatInstanceID is what it is
func (entity *TgChatEntityBase) GetTgChatInstanceID() string {
	return entity.TgChatInstanceID
}

// GetPreferredLanguage returns preferred language for the chat
func (entity *TgChatEntityBase) GetPreferredLanguage() string {
	return entity.PreferredLanguage
}

var _ bots.BotChat = (*TgChatEntityBase)(nil)

// NewTelegramChatEntity create new telegram chat entity
func NewTelegramChatEntity() *TgChatEntityBase {
	return &TgChatEntityBase{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(0, time.Now())},
		},
	}
}

// SetAppUserIntID sets app user int ID
func (entity *TgChatEntityBase) SetAppUserIntID(id int64) {
	if entity.IsGroup && id != 0 {
		panic("TgChatEntityBase.IsGroup && id != 0")
	}
	entity.AppUserIntID = id
}

// SetBotUserID sets bot user int ID
func (entity *TgChatEntityBase) SetBotUserID(id interface{}) {
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

// Load loads entity from datastore
func (entity *TgChatEntityBase) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

// Save saves entity to datastore
func (entity *TgChatEntityBase) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.CleanProperties(properties); err != nil {
		return
	}
	return
}

// CleanProperties cleands properties
func (entity *TgChatEntityBase) CleanProperties(properties []datastore.Property) ([]datastore.Property, error) {
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
			panic(fmt.Sprintf("*TgChatEntityBase.AppUserIntIDs[%d] == 0", i))
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
