package gaehost

import (
	//"fmt"
	"context"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/nds"
	"google.golang.org/appengine/datastore"
	"strconv"
	"time"
	//"reflect"
)

// GaeTelegramChatStore DAL to telegram chat entity
type GaeTelegramChatStore struct {
	GaeBotChatStore
}

var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

// NewGaeTelegramChatStore creates DAL to Telegram chat entity
func NewGaeTelegramChatStore(newTelegramChatEntity func() bots.BotChat) *GaeTelegramChatStore {
	return &GaeTelegramChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore:     NewGaeBaseStore(telegram.ChatKind),
			newBotChatEntity: newTelegramChatEntity,
			validateBotChatEntityType: func(entity bots.BotChat) {
				//if _, ok := entity.(*telegram.TgChatEntityBase); !ok {
				//	v := reflect.ValueOf(entity)
				//	if v.Type() != reflect.TypeOf(telegram.TgChatEntityBase{}) {
				//		panic(fmt.Sprintf("Expected *telegram.TelegramChat but received %T", entity))
				//	}
				//}
			},
			NewBotChatKey: func(c context.Context, botID, botChatId string) *datastore.Key {
				return datastore.NewKey(c, telegram.ChatKind, bots.NewChatID(botID, botChatId), 0, nil)
			},
		},
	}
}

// MarkTelegramChatAsForbidden marks tg chat as forbidden
func MarkTelegramChatAsForbidden(c context.Context, botID string, tgChatID int64, dtForbidden time.Time) error {
	return nds.RunInTransaction(c, func(c context.Context) (err error) {
		key := datastore.NewKey(c, telegram.ChatKind, bots.NewChatID(botID, strconv.FormatInt(tgChatID, 10)), 0, nil)
		var chat telegram.TgChatEntityBase
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
