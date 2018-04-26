package gaehost

import (
	"context"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/viber"
	"github.com/strongo/nds"
	"google.golang.org/appengine/datastore"
	"time"
)

type gaeViberUserChatStore struct {
	GaeBotChatStore
	GaeBotUserStore
}

var _ bots.BotChatStore = (*gaeViberUserChatStore)(nil) // Check for interface implementation at compile time
var _ bots.BotUserStore = (*gaeViberUserChatStore)(nil) // Check for interface implementation at compile time

func newGaeViberUserChatStore(gaeAppUserStore GaeAppUserStore) *gaeViberUserChatStore {
	baseStore := NewGaeBaseStore(viber.UserChatKind)
	return &gaeViberUserChatStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    baseStore,
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				viberUserChatEntity := viber.NewUserChat()
				return &viberUserChatEntity
			},
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*viber.UserChatEntity); !ok {
					panic(fmt.Sprintf("Expected *viber.UserChatEntity but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if stringID, ok := botUserId.(string); ok {
					if stringID == "" {
						panic("botUserKey(): stringID is empty")
					}
					return datastore.NewKey(c, viber.UserChatKind, stringID, 0, nil)
				}
				panic(fmt.Sprintf("Expected botUserId as int, got: %T", botUserId))
			},
		},
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: baseStore,
			newBotChatEntity: func() bots.BotChat {
				viberUserChatEntity := viber.NewUserChat()
				return &viberUserChatEntity
			},
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*viber.UserChatEntity); !ok {
					panic(fmt.Sprintf("Expected *viber.UserChatEntity but received %T", entity))
				}
			},
			NewBotChatKey: func(c context.Context, botID, botChatID string) *datastore.Key {
				return datastore.NewKey(c, viber.UserChatKind, botChatID, 0, nil)
			},
		},
	}
}

// MarkViberChatAsForbidden marks chat as forbidden - TODO: is not used? consider deletion
func MarkViberChatAsForbidden(c context.Context, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, viber.UserChatKind, "", tgChatID, nil)
		var userChatEntity viber.UserChatEntity
		if err = nds.Get(c, key, &userChatEntity); err != nil {
			return
		}
		var changed bool
		if userChatEntity.DtForbidden.IsZero() {
			userChatEntity.DtForbidden = dtForbidden
			changed = true
		}

		if userChatEntity.DtForbiddenLast.IsZero() || userChatEntity.DtForbiddenLast.Before(dtForbidden) {
			userChatEntity.DtForbiddenLast = dtForbidden
			changed = true
		}

		if changed {
			_, err = nds.Put(c, key, &userChatEntity)
		}
		return
	}, nil)
}
