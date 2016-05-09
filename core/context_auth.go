package bots

import (
	"google.golang.org/appengine/log"
	//"google.golang.org/appengine/datastore"
)

func IsAccessGranted(whc WebhookContext) bool {
	return whc.ChatEntity().IsAccessGranted()
}

func SetAccessGranted(whc WebhookContext, value bool) error {
	ctx := whc.Context()
	chatEntity := whc.ChatEntity()
	if chatEntity.IsAccessGranted() == value {
		log.Infof(ctx, "No need to change chatEntity.AccessGranted, as already is: %v", value)
	} else {
		chatEntity.SetAccessGranted(value)
		if err := whc.SaveBotChat(whc.BotChatID(), chatEntity); err != nil {
			return err
		}
	}
	//return SetAccessGrantedForAllUserChats(whc, whc.BotUserKey, value) // TODO: Call in deferrer
	return nil
}

//func SetAccessGrantedForAllUserChats(whc *WebhookContextBase, botUserKey *datastore.Key, value bool) error {
//	//ctx := whc.Context()
//	//var telegramUserEntity TelegramUser
//	//if err := whc.GetOrCreateTelegramUserEntity(&telegramUserEntity); err != nil {
//	//	if err == datastore.ErrNoSuchEntity {
//	//		telegramUserEntity.AccessGranted = !value // We'll update it down the road
//	//	} else {
//	//		return err
//	//	}
//	//}
//	//if telegramUserEntity.AccessGranted == value {
//	//	log.Infof(ctx, "No need to update TelegramUser entity as AccessGranted is already: %v", value)
//	//} else {
//	//	if _, err := SaveTelegramUserEntity(ctx, whc.GetSender().GetID(), &telegramUserEntity); err != nil {
//	//		return err
//	//	}
//	//}
//	//var chats []TelegramChat
//	//chatKeys, err := datastore.NewQuery(TelegramChatKind).Filter("TelegramUserID =", telegramUserID).Filter("AccessGranted =", !value).GetAll(ctx, &chats)
//	//if err != nil {
//	//	return err
//	//}
//	//for i, chat := range chats {
//	//	if chat.AccessGranted != value {
//	//		chatKey, err := SaveTelegramChatEntity(ctx, whc.botSettings.code, chatKeys[i].IntID(), &chat)
//	//		if err != nil {
//	//			log.Warningf(ctx, "Failed to save %v to db", chatKey)
//	//		}
//	//	}
//	//}
//	return nil
//}
//
