package gae_host

import (
	//"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/nds"
	"time"
	"golang.org/x/net/context"
	"strconv"
	//"reflect"
)

type GaeTelegramChatStore struct {
	GaeBotChatStore
}

var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramChatStore(newTelegramChatEntity func() bots.BotChat) *GaeTelegramChatStore {
	return &GaeTelegramChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: NewGaeBaseStore(telegram_bot.TelegramChatKind),
			newBotChatEntity: newTelegramChatEntity,
			validateBotChatEntityType: func(entity bots.BotChat) {
				//if _, ok := entity.(*telegram_bot.TelegramChatEntityBase); !ok {
				//	v := reflect.ValueOf(entity)
				//	if v.Type() != reflect.TypeOf(telegram_bot.TelegramChatEntityBase{}) {
				//		panic(fmt.Sprintf("Expected *telegram_bot.TelegramChat but received %T", entity))
				//	}
				//}
			},
			NewBotChatKey: func(c context.Context, botID, botChatId string) *datastore.Key {
				return datastore.NewKey(c, telegram_bot.TelegramChatKind, bots.NewChatID(botID, botChatId), 0, nil)
			},
		},
	}
}


func MarkTelegramChatAsForbidden(c context.Context, botID string, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, telegram_bot.TelegramChatKind, bots.NewChatID(botID, strconv.FormatInt(tgChatID, 10)), 0, nil)
		var chat telegram_bot.TelegramChatEntityBase
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