package fbm_strongo_bot

import (
	"github.com/strongo/bots-framework/core"
)

const (
	FbmChatKind = "FbmChat"
	FbmUserKind = "FbmUser"
)

type FbmUser struct {
	*bots.BotUserEntity
}

type FbmChat struct {
	*bots.BotChatEntity
	FbmUserID int //TODO: Is it Facebook User ID?
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
