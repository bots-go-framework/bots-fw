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
	BotUserID             int64
	LastProcessedUpdateID int `datastore:",noindex"`
}
var _ bots.BotChat = (*TelegramChat)(nil)

func (chat *TelegramChat) SetAppUserID(id int64) {
	chat.AppUserID = id
}

func(chat *TelegramChat) GetAppUserID() int64 {
	return chat.BotUserID
}

func(chat *TelegramChat) GetBotUserID() int64 {
	return chat.BotUserID
}

func(chat *TelegramChat) SetBotUserID(id int64){
	chat.BotUserID = id
}
