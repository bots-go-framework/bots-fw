package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"time"
	"strconv"
)

const (
	TelegramChatKind = "TgChat"
	TelegramUserKind = "TgUser"
)

type TelegramUserEntity struct {
	bots.BotUserEntity
	//TgChatID int64
}

var _ bots.BotUser = (*TelegramUserEntity)(nil)

type TelegramUser struct {
	ID int64
	TelegramUserEntity
}

func (u TelegramUserEntity) Name() string {
	if u.FirstName == "" && u.LastName == "" {
		return "@" + u.UserName
	}
	name := u.FirstName
	if name != "" {
		name += " " + u.LastName
	} else {
		name = u.LastName
	}
	if u.UserName == "" {
		return name
	}
	return "@" + u.UserName + " - " + name
}

type TelegramChat struct {
	bots.BotChatEntity
	TelegramUserID        int
	LastProcessedUpdateID int `datastore:",noindex"`
}

var _ bots.BotChat = (*TelegramChat)(nil)

func NewTelegramChat() TelegramChat {
	return TelegramChat{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: bots.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}

func (chat *TelegramChat) SetAppUserIntID(id int64) {
	chat.AppUserIntID = id
}

//func (chat *TelegramChat) GetBotUserIntID() int {
//	return chat.TelegramUserID
//}

func (chat *TelegramChat) SetBotUserID(id interface{}) {
	switch id.(type) {
	case string:
		var err error
		chat.TelegramUserID, err = strconv.Atoi(id.(string))
		if err != nil {
			panic(err.Error())
		}
	case int:
		chat.TelegramUserID = id.(int)
	case int64:
		chat.TelegramUserID = id.(int)
	default:
		panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
	}
}
