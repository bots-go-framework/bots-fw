package telegram_bot

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"time"
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
	DtForbidden time.Time
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

func (chat *TelegramChat) GetBotUserID() interface{} {
	return chat.TelegramUserID
}

func (chat *TelegramChat) SetBotUserID(id interface{}) {
	if intId, ok := id.(int); ok {
		chat.TelegramUserID = intId
		return
	}
	if intId64, ok := id.(int64); ok {
		chat.TelegramUserID = int(intId64)
		return
	}
	panic(fmt.Sprintf("Expected int, got: %T", id))

}
