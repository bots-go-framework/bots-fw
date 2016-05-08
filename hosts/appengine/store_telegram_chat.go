package gae_host

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"fmt"
	"google.golang.org/appengine/log"
)

type GaeTelegramChatStore struct {
	GaeBotChatStore
}
var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramChatStore(c context.Context) *GaeTelegramChatStore {
	return &GaeTelegramChatStore{
		GaeBotChatStore: GaeBotChatStore{
			GaeBaseStore: GaeBaseStore{c: c, entityKind: telegram_bot.TelegramChatKind},
			newBotChatEntity: func() bots.BotChat { return &telegram_bot.TelegramChat{} },
			validateBotChatEntityType: func(entity bots.BotChat) {
				if _, ok := entity.(*telegram_bot.TelegramChat); ok {
					panic(fmt.Sprintf("Expected *telegram_bot.TelegramChat but received %t", entity))
				}
			},
			botChatKey: func(botChatId interface{}) *datastore.Key {
				if intId, ok := botChatId.(int64); ok {
					key := datastore.NewKey(c, telegram_bot.TelegramChatKind, "", intId, nil)
					log.Infof(c, "BotChatKey: %v", key)
					return key
				} else {
					panic(fmt.Sprintf("Expected botChatId as int64, got: %t", botChatId))
				}
			},
		},
	}
}