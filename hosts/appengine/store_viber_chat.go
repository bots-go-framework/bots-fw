package gae_host

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
	"github.com/qedus/nds"
	"time"
	"golang.org/x/net/context"
	"github.com/strongo/bots-framework/platforms/viber"
)

type GaeViberChatStore struct {
	GaeBotChatStore
}

var _ bots.BotChatStore = (*GaeViberChatStore)(nil) // Check for interface implementation at compile time

func NewGaeViberChatStore(log strongo.Logger, r *http.Request) *GaeViberChatStore {
	return &GaeViberChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: NewGaeBaseStore(log, r, viber_bot.ViberChatKind),
			newBotChatEntity: func() bots.BotChat {
				telegramChat := viber_bot.NewViberChat()
				return &telegramChat
			},
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*viber_bot.ViberChat); !ok {
					panic(fmt.Sprintf("Expected *viber_bot.ViberChat but received %T", entity))
				}
			},
			botChatKey: func(botChatId interface{}) *datastore.Key {
				if intId, ok := botChatId.(int64); ok {
					key := datastore.NewKey(appengine.NewContext(r), viber_bot.ViberChatKind, "", (int64)(intId), nil)
					return key
				} else if strId, ok := botChatId.(string); ok {
					key := datastore.NewKey(appengine.NewContext(r), viber_bot.ViberChatKind, strId, 0, nil)
					return key
				} else {
					panic(fmt.Sprintf("Expected botChatId as int, got: %T", botChatId))
				}
			},
		},
	}
}


func MarkViberChatAsForbidden(c context.Context, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, viber_bot.ViberChatKind, "", tgChatID, nil)
		var chat viber_bot.ViberChat
		if err = nds.Get(c, key, &chat); err != nil {
			return
		}
		var changed bool
		if chat.DtForbidden.IsZero() {
			chat.DtForbidden = dtForbidden
			changed = true
		}

		if chat.DtForbiddenLast.IsZero() || chat.DtForbiddenLast.Before(dtForbidden) {
			chat.DtForbiddenLast = dtForbidden
			changed = true
		}

		if changed {
			_, err = nds.Put(c, key, &chat)
		}
		return
	}, nil)
}