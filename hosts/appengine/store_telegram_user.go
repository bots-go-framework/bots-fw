package gae_host

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
	"time"
	"golang.org/x/net/context"
	"github.com/strongo/app/user"
)

type GaeTelegramUserStore struct {
	GaeBotUserStore
}

var _ bots.BotUserStore = (*GaeTelegramUserStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramUserStore(gaeAppUserStore GaeAppUserStore) GaeTelegramUserStore {
	return GaeTelegramUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    NewGaeBaseStore(telegram_bot.TelegramUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				if apiUser == nil {
					return &telegram_bot.TelegramUserEntity{}
				} else {
					return &telegram_bot.TelegramUserEntity{
						BotUserEntity: bots.BotUserEntity{
							BotEntity: bots.BotEntity{
								OwnedByUser: user.OwnedByUser{
									DtCreated: time.Now(),
								},
							},
							FirstName: apiUser.GetFirstName(),
							LastName:  apiUser.GetLastName(),
							UserName:  apiUser.GetUserName(),
						},
					}
				}
			},
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*telegram_bot.TelegramUserEntity); !ok {
					panic(fmt.Sprintf("Expected *telegram_bot.TelegramUser but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if intID, ok := botUserId.(int); ok {
					if intID == 0 {
						panic("botUserKey(): intID == 0")
					}
					return datastore.NewKey(c, telegram_bot.TelegramUserKind, "", (int64)(intID), nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as int, got: %T", botUserId))
				}
			},
		},
	}
}
