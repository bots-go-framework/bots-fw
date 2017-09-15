package telegram_bot

import (
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/app/user"
)

const (
	TelegramUserKind = "TgUser"
)

type TelegramUserEntity struct {
	bots.BotUserEntity
	//TgChatID int64
}

var _ bots.BotUser = (*TelegramUserEntity)(nil)
var _ user.AccountEntity = (*TelegramUserEntity)(nil)

type TelegramUser struct {
	ID int64
	TelegramUserEntity
}

func (_ TelegramUser) GetEmail() string {
	return ""
}

func (entity TelegramUserEntity) Name() string {
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

func (entity *TelegramUserEntity) GetNames() user.Names {
	return user.Names{
		FirstName: entity.FirstName,
		LastName: entity.LastName,
		NickName: entity.UserName,
	}
}

func (entity *TelegramUserEntity) IsEmailConfirmed() bool {
	return false
}

func (entity *TelegramUserEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *TelegramUserEntity) Save() (properties []datastore.Property, err error) {
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
