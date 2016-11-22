package fbm_bot

import (
	"github.com/strongo/bots-framework/core"
	"time"
	"fmt"
)

const (
	FbmChatKind = "FbmChat"
	FbmUserKind = "FbmUser"
)

type FbmUser struct {
	bots.BotUserEntity
}

type FbmChat struct {
	bots.BotChatEntity
	FbmUserID string //TODO: Is it Facebook User ID?
	LastSeq int
}

func (chat *FbmChat) SetBotUserID(id interface{}) {
	switch id.(type) {
	case string:
		chat.FbmUserID = id.(string)
	default:
		panic(fmt.Sprintf("Expected string, got: %T=%v", id, id))
	}
}

func (chat *FbmChat) GetBotUserStringID() string {
	return chat.FbmUserID
}


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

func NewFbmChat() FbmChat {
	return FbmChat{
		BotChatEntity: bots.BotChatEntity{
			BotEntity: bots.BotEntity{
				OwnedByUser: bots.OwnedByUser{
					DtCreated: time.Now(),
				},
			},
		},
	}
}
