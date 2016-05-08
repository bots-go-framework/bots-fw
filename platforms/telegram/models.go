package telegram_bot

import (
	"github.com/strongo/bots-framework/core"
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
	TelegramUserID             int64
	LastProcessedUpdateID int `datastore:",noindex"`
}
var _ bots.BotChat = (*TelegramChat)(nil)

func (chat *TelegramChat) SetAppUserID(id int64) {
	chat.UserID = id
}

func(chat *TelegramChat) GetAppUserID() int64 {
	return chat.TelegramUserID
}

func(chat *TelegramChat) GetBotUserID() int64 {
	return chat.TelegramUserID
}

func(chat *TelegramChat) SetBotUserID(id int64){
	chat.TelegramUserID = id
}
