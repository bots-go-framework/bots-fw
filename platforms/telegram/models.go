package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"fmt"
	"time"
)

const (
	TelegramChatKind = "TgChat"
	TelegramUserKind = "TgUser"
)

type TelegramUser struct {
	bots.BotUserEntity
}
var _ bots.BotUser = (*TelegramUser)(nil)

type TelegramChat struct {
	bots.BotChatEntity
	TelegramUserID             int
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

func (chat *TelegramChat) SetAppUserID(id int64) {
	chat.AppUserID = id
}

func(chat *TelegramChat) GetBotUserID() interface{} {
	return chat.TelegramUserID
}

func(chat *TelegramChat) SetBotUserID(id interface{}){
	if intId, ok := id.(int); ok {
		chat.TelegramUserID = intId
	} else {
		panic(fmt.Sprintf("Expected int, got: %T", id))
	}
}

