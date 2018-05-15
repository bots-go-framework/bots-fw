package gaehost

import (
	"context"
	"fmt"
	"github.com/strongo/app/user"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"google.golang.org/appengine/datastore"
	"time"
)

type gaeFacebookUserStore struct {
	GaeBotUserStore
}

var _ bots.BotUserStore = (*gaeFacebookUserStore)(nil) // Check for interface implementation at compile time

func newGaeFacebookUserStore(gaeAppUserStore GaeAppUserStore) gaeFacebookUserStore {
	return gaeFacebookUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    NewGaeBaseStore(fbm.BotUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				if apiUser == nil {
					return &fbm.BotUser{}
				}
				return &fbm.BotUser{
					BotUserEntity: bots.BotUserEntity{
						BotEntity: bots.BotEntity{OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(0, time.Now())},
						FirstName: apiUser.GetFirstName(),
						LastName:  apiUser.GetLastName(),
						UserName:  apiUser.GetUserName(),
					},
				}
			},
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*fbm.BotUser); !ok {
					panic(fmt.Sprintf("Expected *fbm.BotUser but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if stringID, ok := botUserId.(string); ok {
					return datastore.NewKey(c, fbm.BotUserKind, stringID, 0, nil)
				}
				panic(fmt.Sprintf("Expected botUserId as string, got: %T", botUserId))
			},
		},
	}
}
