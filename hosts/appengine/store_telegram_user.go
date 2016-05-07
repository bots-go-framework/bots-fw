package gae_host

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"fmt"
)

type GaeTelegramUserStore struct {
	GaeBotUserStore
}
var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramUserStore(c context.Context) GaeTelegramUserStore {
	return GaeTelegramUserStore{
		GaeBotUserStore: GaeBotUserStore{
			GaeBaseStore: GaeBaseStore{c: c, entityKind: telegram_bot.TelegramUserKind},
			newBotUserEntity: func() bots.BotUser { return &telegram_bot.TelegramChat{} },
			validateBotUserEntityType: func(entity bots.BotUser) {
				if _, ok := entity.(*telegram_bot.TelegramUser); !ok {
					panic(fmt.Sprintf("Expected *telegram_bot.TelegramUser but received %t", entity))
				}
			},
			botUserKey: func(botUserId interface{}) *datastore.Key {
				if intID, ok := botUserId.(int64); ok {
					return datastore.NewKey(c, telegram_bot.TelegramUserKind, "", intID, nil)
				} else {
					panic(fmt.Sprintf("Expected botUserId as int64, got: %t", botUserId))
				}
			},
		},
	}
}