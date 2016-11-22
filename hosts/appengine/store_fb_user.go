package gae_host

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"time"
	"golang.org/x/net/context"
	"github.com/strongo/bots-framework/platforms/fbm"
)

type GaeFacebookUserStore struct {
	GaeBotUserStore
}

var _ bots.BotUserStore = (*GaeFacebookUserStore)(nil) // Check for interface implementation at compile time

func NewGaeFacebookUserStore(log strongo.Logger, gaeAppUserStore GaeAppUserStore) GaeFacebookUserStore {
	return GaeFacebookUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    NewGaeBaseStore(log, fbm_bot.FbmUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				if apiUser == nil {
					return &fbm_bot.FbmUser{}
				} else {
					return &fbm_bot.FbmUser{
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
				if _, ok := entity.(*fbm_bot.FbmUser); !ok {
					panic(fmt.Sprintf("Expected *fbm_bot.FbmUser but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if stringID, ok := botUserId.(string); ok {
					return datastore.NewKey(c, fbm_bot.FbmUserKind, stringID, 0, nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as string, got: %T", botUserId))
				}
			},
		},
	}
}
