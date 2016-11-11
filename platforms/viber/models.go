package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"time"
	"fmt"
)

const (
	ViberChatKind = "ViberChat"
	ViberUserKind = "ViberUser"
)

type ViberUser struct {
	bots.BotUserEntity
	//TgChatID int64
}

var _ bots.BotUser = (*ViberUser)(nil)

type ViberChat struct {
	bots.BotChatEntity
	ViberUserID        string
}

var _ bots.BotChat = (*ViberChat)(nil)

func NewViberChat() ViberChat {
	return ViberChat{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: bots.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}

func (chat *ViberChat) SetAppUserIntID(id int64) {
	chat.AppUserIntID = id
}

func (chat *ViberChat) GetBotUserStringID() string {
	return chat.ViberUserID
}

func (chat *ViberChat) SetBotUserID(id interface{}) {
	if intId, ok := id.(string); ok {
		chat.ViberUserID = intId
		return
	}
	panic(fmt.Sprintf("Expected string, got: %T", id))

}
