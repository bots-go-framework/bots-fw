package telegram

import (
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

const (
	// TgUserKind is kind name for Telegram user entity
	TgUserKind = "TgUser"
)

// TgUserEntity is Telegram user DB entity (without ID)
type TgUserEntity struct {
	bots.BotUserEntity
	//TgChatID int64
}

var _ bots.BotUser = (*TgUserEntity)(nil)
var _ user.AccountEntity = (*TgUserEntity)(nil)

// TgUser is Telegram user DB record (with ID)
type TgUser struct {
	ID int64
	TgUserEntity
}

// GetEmail returns empty string
func (TgUser) GetEmail() string {
	return ""
}

// Name returns full display name cmbined from (first+last, nick) name
func (entity TgUserEntity) Name() string {
	if entity.FirstName == "" && entity.LastName == "" {
		return "@" + entity.UserName
	}
	name := entity.FirstName
	if name != "" {
		name += " " + entity.LastName
	} else {
		name = entity.LastName
	}
	if entity.UserName == "" {
		return name
	}
	return "@" + entity.UserName + " - " + name
}

// GetNames return user names
func (entity *TgUserEntity) GetNames() user.Names {
	return user.Names{
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		NickName:  entity.UserName,
	}
}

// IsEmailConfirmed returns false
func (entity *TgUserEntity) IsEmailConfirmed() bool {
	return false
}

// Load is for datastore
func (entity *TgUserEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

// Save is for datastore
func (entity *TgUserEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}

	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"AccessGranted": gaedb.IsFalse,
		"FirstName":     gaedb.IsEmptyString,
		"LastName":      gaedb.IsEmptyString,
		"UserName":      gaedb.IsEmptyString,
	}); err != nil {
		return
	}

	return
}
