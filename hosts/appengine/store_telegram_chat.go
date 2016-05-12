package gae_host

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"fmt"
	"net/http"
	"google.golang.org/appengine"
)

type GaeTelegramChatStore struct {
	GaeBotChatStore
}
var _ bots.BotChatStore = (*GaeTelegramChatStore)(nil) // Check for interface implementation at compile time

func NewGaeTelegramChatStore(log bots.Logger, r *http.Request) *GaeTelegramChatStore {
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
				if intId, ok := botChatId.(int); ok {
					key := datastore.NewKey(appengine.NewContext(r), telegram_bot.TelegramChatKind, "", (int64)(intId), nil)
					log.Infof("BotChatKey: %v", key)
					return key
				} else {
					panic(fmt.Sprintf("Expected botChatId as int, got: %T", botChatId))
				}
			},
		},
	}
}