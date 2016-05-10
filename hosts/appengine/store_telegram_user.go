package gae_host

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"fmt"
	"net/http"
	"google.golang.org/appengine"
)

type GaeTelegramUserStore struct {
	GaeBotUserStore
}
var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramUserStore(log bots.Logger, r *http.Request, gaeAppUserStore GaeAppUserStore) GaeTelegramUserStore {
	return GaeTelegramUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore: NewGaeBaseStore(log, r, telegram_bot.TelegramUserKind),
			gaeAppUserStore: gaeAppUserStore,
			newBotUserEntity: func() bots.BotUser { return &telegram_bot.TelegramChat{} },
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*telegram_bot.TelegramUser); !ok {
					panic(fmt.Sprintf("Expected *telegram_bot.TelegramUser but received %T", entity))
				}
			},
			botUserKey: func(botUserId interface{}) *datastore.Key {
				if intID, ok := botUserId.(int); ok {
					return datastore.NewKey(appengine.NewContext(r), telegram_bot.TelegramUserKind, "", (int64)(intID), nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as int, got: %T", botUserId))
				}
			},
		},
	}
}