package gae_host

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
	"github.com/qedus/nds"
	"time"
	"golang.org/x/net/context"
)

type GaeTelegramChatStore struct {
	GaeBotChatStore
}

var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramChatStore(log strongo.Logger, r *http.Request) *GaeTelegramChatStore {
	return &GaeTelegramChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: NewGaeBaseStore(log, r, telegram_bot.TelegramChatKind),
			newBotChatEntity: func() bots.BotChat {
				telegramChat := telegram_bot.NewTelegramChat()
				return &telegramChat
			},
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*telegram_bot.TelegramChat); !ok {
					panic(fmt.Sprintf("Expected *telegram_bot.TelegramChat but received %T", entity))
				}
			},
			botChatKey: func(botChatId interface{}) *datastore.Key {
				if intId, ok := botChatId.(int64); ok {
					key := datastore.NewKey(appengine.NewContext(r), telegram_bot.TelegramChatKind, "", (int64)(intId), nil)
					return key
				} else if strId, ok := botChatId.(string); ok {
					key := datastore.NewKey(appengine.NewContext(r), telegram_bot.TelegramChatKind, strId, 0, nil)
					return key
				} else {
					panic(fmt.Sprintf("Expected botChatId as int, got: %T", botChatId))
				}
			},
		},
	}
}


func MarkTelegramChatAsForbidden(c context.Context, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, telegram_bot.TelegramChatKind, "", tgChatID, nil)
		var chat telegram_bot.TelegramChat
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