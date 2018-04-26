package fbm

import (
	"fmt"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"time"
)

const (
	// ChatKind is kind name for FBM chat entity
	ChatKind = "Chat"

	// BotUserKind is kind name for FBM user entity
	BotUserKind = "FbmUser"
)

// BotUser entity
type BotUser struct {
	bots.BotUserEntity
}

// Chat entity
type Chat struct {
	bots.BotChatEntity
	FbmUserID string //TODO: Is it Facebook User ID?
	LastSeq   int
}

// SetBotUserID sets bot user ID
func (chat *Chat) SetBotUserID(id interface{}) {
	switch id.(type) {
	case string:
		chat.FbmUserID = id.(string)
	default:
		panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
	}
}

//func (chat *Chat) GetBotUserStringID() string {
//	return chat.FbmUserID
//}

//func GetUserByFbmUserID(ctx context.Context, telegramUserID int, createIfMissing bool) (*datastore.Key, *common.User, error) {
//	botUser := bot.BotUser{}
//	err := GetTelegramUserEntity(ctx, telegramUserID, &botUser)
//	if err == nil {
//		userKey := datastore.NewKey(ctx, common.UserKind, "", botUser.UserID, nil)
//		user := common.User{}
//		err = nds.Get(ctx, userKey, &user)
//		return userKey, &user, err
//	}
//	return nil, nil, err
//}

// NewFbmChat create new FBM chat entity
func NewFbmChat() Chat {
	return Chat{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: user.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}
