package gae_host

import (
	"context"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/viber"
	"github.com/strongo/nds"
	"google.golang.org/appengine/datastore"
	"time"
)

type GaeViberUserChatStore struct {
	GaeBotChatStore
	GaeBotUserStore
}

var _ bots.BotChatStore = (*GaeViberUserChatStore)(nil) // Check for interface implementation at compile time
var _ bots.BotUserStore = (*GaeViberUserChatStore)(nil) // Check for interface implementation at compile time

func NewGaeViberUserChatStore(gaeAppUserStore GaeAppUserStore) *GaeViberUserChatStore {
	baseStore := NewGaeBaseStore(viber_bot.ViberUserChatKind)
	return &GaeViberUserChatStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore:    baseStore,
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func(apiUser bots.WebhookActor) bots.BotUser {
				viberUserChatEntity := viber_bot.NewViberUserChat()
				return &viberUserChatEntity
			},
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*viber_bot.ViberUserChatEntity); !ok {
					panic(fmt.Sprintf("Expected *viber_bot.ViberUserChatEntity but received %T", entity))
				}
			},
			botUserKey: func(c context.Context, botUserId interface{}) *datastore.Key {
				if stringID, ok := botUserId.(string); ok {
					if stringID == "" {
						panic("botUserKey(): stringID is empty")
					}
					return datastore.NewKey(c, viber_bot.ViberUserChatKind, stringID, 0, nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as int, got: %T", botUserId))
				}
			},
		},
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: baseStore,
			newBotChatEntity: func() bots.BotChat {
				viberUserChatEntity := viber_bot.NewViberUserChat()
				return &viberUserChatEntity
			},
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*viber_bot.ViberUserChatEntity); !ok {
					panic(fmt.Sprintf("Expected *viber_bot.ViberUserChatEntity but received %T", entity))
				}
			},
			NewBotChatKey: func(c context.Context, botID, botChatID string) *datastore.Key {
				return datastore.NewKey(c, viber_bot.ViberUserChatKind, botChatID, 0, nil)
			},
		},
	}
}

func MarkViberChatAsForbidden(c context.Context, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, viber_bot.ViberUserChatKind, "", tgChatID, nil)
		var userChatEntity viber_bot.ViberUserChatEntity
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
