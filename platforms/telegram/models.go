package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
	"fmt"
)

const (
	TelegramChatKind = "TgChat"
	TelegramUserKind = "TgUser"
)

type TelegramUser struct {
	bots.BotUserEntity
}

type TelegramChat struct {
	bots.BotChatEntity
	TelegramUserID             int
	LastProcessedUpdateID int `datastore:",noindex"`
}
var _ bots.BotChat = (*TelegramChat)(nil)

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
