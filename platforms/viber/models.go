package viber

import (
	"fmt"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"time"
)

const (
	//ViberChatKind = "ViberChat"
	//ViberUserKind = "ViberUser"

	// UserChatKind is user chat kind name
	UserChatKind = "ViberUserChat"
)

// UserChatEntity is bot chat entity for Viber
type UserChatEntity struct {
	bots.BotChatEntity
	//ViberUserID string `datastore:",noindex"` // Duplicate of key.StringID(), required for GetBotUserStringID()
	//UserName string `datastore:",noindex"`
	//Avatar string `datastore:",noindex"`
}

var _ bots.BotUser = (*UserChatEntity)(nil)
var _ bots.BotChat = (*UserChatEntity)(nil)

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

// NewUserChat creates new UserChatEntity
func NewUserChat() UserChatEntity {
	return UserChatEntity{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(0, time.Now())},
		},
	}
}

// SetAppUserIntID sets app user int ID
func (chat *UserChatEntity) SetAppUserIntID(id int64) {
	chat.AppUserIntID = id
}

//func (chat *ViberChat) GetBotUserStringID() string {
//	return chat.ViberUserID
//}

// SetBotUserID sets bot user ID
func (chat *UserChatEntity) SetBotUserID(id interface{}) {
	if _, ok := id.(string); ok {
		// Ignore as stored in the key. chat.ViberUserID = stringID
		return
	}
	panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
}
