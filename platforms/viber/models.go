package viber_bot

import (
	"github.com/strongo/bots-framework/core"
	"time"
	"fmt"
	"github.com/strongo/app/user"
)

const (
	//ViberChatKind = "ViberChat"
	//ViberUserKind = "ViberUser"
	ViberUserChatKind = "ViberUserChat"
)

type ViberUserChatEntity struct {
	bots.BotChatEntity
	//ViberUserID string `datastore:",noindex"` // Duplicate of key.StringID(), required for GetBotUserStringID()
	//UserName string `datastore:",noindex"`
	//Avatar string `datastore:",noindex"`
}
var _ bots.BotUser = (*ViberUserChatEntity)(nil)
var _ bots.BotChat = (*ViberUserChatEntity)(nil)

//type ViberUser struct { //TODO: Get rid of the entity. Move props like Name to ViberChat entity.
//	bots.BotUserEntity
//	//TgChatID int64
//}


//var _ bots.BotUser = (*ViberUser)(nil)

//type ViberChat struct {
//	bots.BotChatEntity
//	ViberUserID        string
//}
//
//var _ bots.BotChat = (*ViberChat)(nil)

func NewViberUserChat() ViberUserChatEntity {
	return ViberUserChatEntity{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: user.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}

func (chat *ViberUserChatEntity) SetAppUserIntID(id int64) {
	chat.AppUserIntID = id
}

//func (chat *ViberChat) GetBotUserStringID() string {
//	return chat.ViberUserID
//}

func (chat *ViberUserChatEntity) SetBotUserID(id interface{}) {
	if _, ok := id.(string); ok {
		// Ignore as stored in the key. chat.ViberUserID = stringID
		return
	}
	panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
}
