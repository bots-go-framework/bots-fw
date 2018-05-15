package gaehost

import (
	"context"
	"fmt"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
	"time"
)

// gaeTelegramUserStore is DAL to telegram user entity
type gaeTelegramUserStore struct {
	GaeBotUserStore
}

var _ bots.BotUserStore = (*gaeTelegramUserStore)(nil) // Check for interface implementation at compile time

// newGaeTelegramUserStore create DAL to Telegram user entity
func newGaeTelegramUserStore(gaeAppUserStore GaeAppUserStore) gaeTelegramUserStore {
	return gaeTelegramUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    NewGaeBaseStore(telegram.TgUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				if apiUser == nil {
					return &telegram.TgUserEntity{}
				}
				botEntity := bots.BotEntity{
					OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(0, time.Now()),
				}
				botUserEntity := bots.BotUserEntity{
					BotEntity: botEntity,
					FirstName: apiUser.GetFirstName(),
					LastName:  apiUser.GetLastName(),
					UserName:  apiUser.GetUserName(),
				}
				return &telegram.TgUserEntity{BotUserEntity: botUserEntity}
			},
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*telegram.TgUserEntity); !ok {
					panic(fmt.Sprintf("Expected *telegram.TgUser but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if intID, ok := botUserId.(int); ok {
					if intID == 0 {
						panic("botUserKey(): intID == 0")
					}
					return datastore.NewKey(c, telegram.TgUserKind, "", (int64)(intID), nil)
				}
				panic(fmt.Sprintf("Expected botUserId as int, got: %T", botUserId))
			},
		},
	}
}
