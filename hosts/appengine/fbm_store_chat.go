package gae_host

import (
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type GaeFbmChatStore struct {
	GaeBotChatStore
}

var _ bots.BotChatStore = (*GaeFbmChatStore)(nil) // Check for interface implementation at compile time

func NewGaeFbmChatStore() *GaeTelegramChatStore {
	return &GaeTelegramChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: NewGaeBaseStore(fbm_bot.FbmChatKind),
			newBotChatEntity: func() bots.BotChat {
				telegramChat := fbm_bot.NewFbmChat()
				return &telegramChat
			},
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*fbm_bot.FbmChat); !ok {
					panic(fmt.Sprintf("Expected *fbm_bot.FbmChat but received %T", entity))
				}
			},
			NewBotChatKey: func(c context.Context, botID, botChatId string) *datastore.Key {
				return datastore.NewKey(c, fbm_bot.FbmChatKind, bots.NewChatID(botID, botChatId), 0, nil)
			},
		},
	}
}

//func MarkFacebookChatAsForbidden(c context.Context, botID string, tgChatID int64, dtForbidden time.Time) error {
//	return nds.RunInTransaction(c, func(c context.Context) (err error) {
//		key := datastore.NewKey(c, telegram_bot.TelegramChatKind, bots.NewChatID(botID, strconv.FormatInt(tgChatID, 10)), 0, nil)
//		var chat telegram_bot.TelegramChat
//		if err = nds.Get(c, key, &chat); err != nil {
//			return
//		}
//		var changed bool
//		if chat.DtForbidden.IsZero() {
//			chat.DtForbidden = dtForbidden
//			changed = true
//		}
//
//		if chat.DtForbiddenLast.IsZero() || chat.DtForbiddenLast.Before(dtForbidden) {
//			chat.DtForbiddenLast = dtForbidden
//			changed = true
//		}
//
//		if changed {
//			_, err = nds.Put(c, key, &chat)
//		}
//		return
//	}, nil)
//}
