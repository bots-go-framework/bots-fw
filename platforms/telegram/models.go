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

type TelegramUser struct {
	bots.BotUserEntity
	//TgChatID int64
}

var _ bots.BotUser = (*TelegramUser)(nil)

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

func (chat *TelegramChat) GetBotUserIntID() int {
	return chat.TelegramUserID
}

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
		panic(fmt.Sprintf("Expected int or string, got: %T", id))
	}
}
