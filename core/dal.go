package bots

import (
	"github.com/qedus/nds"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
)


func LoadBotChatEntity(c context.Context, botChatKey *datastore.Key, entity BotChat) (err error) {
	return nds.Get(c, botChatKey, entity)
}

func SaveBotChatEntity(c context.Context, botChatKey *datastore.Key, entity BotChat) (*datastore.Key, error) {
	entity.SetDtUpdatedToNow()
	return nds.Put(c, botChatKey, entity)
}

func GetBotUserEntity(c context.Context, botUserKey *datastore.Key, entity BotUser) (err error) {
	return nds.Get(c, botUserKey, entity)
}

//func SaveBotUserEntity(c context.Context, id int, entity *TelegramUser) (*datastore.Key, error) {
//	return nds.Put(c, NewTelegramUserEntityKey(c, id), entity)
//}

