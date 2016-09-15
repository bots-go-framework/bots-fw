package bots

import (
	"google.golang.org/appengine/log"
	//"google.golang.org/appengine/datastore"
	"github.com/pkg/errors"
)

func IsAccessGranted(whc WebhookContext) bool {
	return whc.ChatEntity().IsAccessGranted()
}

func SetAccessGranted(whc WebhookContext, value bool) (err error) {
	logger := whc.Logger()
	logger.Debugf("SetAccessGranted(value=%v)", value)
	ctx := whc.Context()
	chatEntity := whc.ChatEntity()
	if chatEntity != nil {
		if chatEntity.IsAccessGranted() == value {
			log.Infof(ctx, "No need to change chatEntity.AccessGranted, as already is: %v", value)
		} else {
			chatEntity.SetAccessGranted(value)
			if err := whc.SaveBotChat(whc.BotChatID(), chatEntity); err != nil {
				return errors.Wrap(err, "Failed to save bot chat entity to db")
			}
		}
	}

	botUserID := whc.GetSender().GetID()
	logger.Debugf("SetAccessGranted(): whc.GetSender().GetID() = %v", botUserID)
	if botUser, err := whc.GetBotUserById(botUserID); err != nil {
		return errors.Wrapf(err, "Failed to get bot user by id=%v", botUserID)
	} else {
		botUser.SetAccessGranted(value)
		if err = whc.SaveBotUser(botUserID, botUser); err != nil {
			err = errors.Wrapf(err, "Failed to call whc.SaveBotUser(botUserID=%v)", botUserID)
		}
		return err
	}
	//return SetAccessGrantedForAllUserChats(whc, whc.BotUserKey, value) // TODO: Call in deferrer
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
