package telegram_bot

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/cacheddb"
	"strconv"
)

const (
	TelegramChatKind = "TgChat";
	TelegramUserKind = "TgUser";
)

type TelegramUser struct {
	bots.BotUserEntity
}

//func TelegramUserIdFromTelegramUserKey(key *datastore.Key) int64 {
//	if key.Kind() != TelegramUserKind {
//		panic(fmt.Sprintf("Invalid key, got kind %v, expected %v", key.Kind(), TelegramUserKind))
//	}
//	return key.IntID()
//}

type TelegramChat struct {
	bots.BotChatEntity
	TelegramUserID int
	LastProcessedUpdateID int `datastore:",noindex"`
}

func (e *TelegramChat) SetUserID(id int64) {
	e.UserID = id
}

func GetUserByTelegramID(ctx context.Context, telegramUserID int, createIfMissing bool) (*datastore.Key, bots.UserEntity, error) {
	telegramUser := TelegramUser{}
	err := bots.GetBotUserEntity(ctx, NewTelegramUserEntityKey(ctx, telegramUserID), &telegramUser)
	if err != nil {
		return nil, nil, err
	}
	userKey := datastore.NewKey(ctx, common.UserKind, "", telegramUser.UserID, nil)
	user := common.User{}
	err = cacheddb.Get(ctx, userKey, &user)
	return userKey, &user, err
}

func NewTelegramChatEntityKey(c context.Context, botID string, chatID int) *datastore.Key {
	return datastore.NewKey(c, TelegramChatKind, botID + strconv.FormatInt(int64(chatID), 10), 0, nil)
}

func NewTelegramUserEntityKey(c context.Context, id int) *datastore.Key {
	return datastore.NewKey(c, TelegramUserKind, "", int64(id), nil)
}

