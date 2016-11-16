package gae_host

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/viber"
	"google.golang.org/appengine/datastore"
	"time"
	"golang.org/x/net/context"
)

type GaeViberUserStore struct {
	GaeBotUserStore
}

var _ bots.BotUserStore = (*GaeViberUserStore)(nil) // Check for interface implementation at compile time

func NewGaeViberUserStore(log strongo.Logger, gaeAppUserStore GaeAppUserStore) GaeViberUserStore {
	return GaeViberUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    NewGaeBaseStore(log, viber_bot.ViberUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				if apiUser == nil {
					return &viber_bot.ViberUser{}
				} else {
					return &viber_bot.ViberUser{
						BotUserEntity: bots.BotUserEntity{
							BotEntity: bots.BotEntity{
								OwnedByUser: bots.OwnedByUser{
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
				if _, ok := entity.(*viber_bot.ViberUser); !ok {
					panic(fmt.Sprintf("Expected *viber_bot.ViberUser but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if stringID, ok := botUserId.(string); ok {
					if stringID == "" {
						panic("botUserKey(): botUserId is an empty string ")
					}
					return datastore.NewKey(c, viber_bot.ViberUserKind, stringID, 0, nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as string, got: %T:%v", botUserId, botUserId))
				}
			},
		},
	}
}
