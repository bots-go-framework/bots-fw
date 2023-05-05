package botsfw

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

// SetAccessGranted marks current context as authenticated
func SetAccessGranted(whc WebhookContext, value bool) (err error) {
	c := whc.Context()
	log.Debugf(c, "SetAccessGranted(value=%v)", value)
	chatEntity := whc.ChatEntity()
	if chatEntity != nil {
		if chatEntity.IsAccessGranted() == value {
			log.Infof(c, "No need to change chatEntity.AccessGranted, as already is: %v", value)
		} else {
			if err = whc.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
				var chatID string
				if chatID, err = whc.BotChatID(); err != nil {
					return
				}
				if chatEntity, err = whc.GetBotChatEntityByID(c, whc.GetBotCode(), chatID); err != nil {
					return
				}
				if changed := chatEntity.SetAccessGranted(value); changed {
					if err = whc.SaveBotChat(c, whc.GetBotCode(), chatID, chatEntity); err != nil {
						err = fmt.Errorf("failed to save bot chat entity to db: %w", err)
					}
				}
				return
			}, nil); err != nil {
				return
			}
		}
	}

	botUserID := whc.GetSender().GetID()
	botUserStrID := fmt.Sprintf("%v", botUserID)
	log.Debugf(c, "SetAccessGranted(): whc.GetSender().GetID() = %v", botUserID)
	if botUser, err := whc.GetBotUserByID(c, botUserStrID); err != nil {
		return fmt.Errorf("failed to get bot user by id=%v: %w", botUserID, err)
	} else if botUser.IsAccessGranted() == value {
		log.Infof(c, "No need to change botUser.AccessGranted, as already is: %v", value)
	} else {
		err = whc.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
			botUser.SetAccessGranted(value)
			if botUser, err = whc.GetBotUserByID(c, botUserStrID); err != nil {
				return fmt.Errorf("failed to get transactionally bot user by id=%v: %w", botUserID, err)
			}
			if changed := botUser.SetAccessGranted(value); changed {
				if err = whc.SaveBotUser(c, botUserStrID, botUser); err != nil {
					err = fmt.Errorf("failed to call whc.SaveBotUser(botUserID=%v): %w", botUserID, err)
				}
			}
			return err
		}, nil)
	}
	return err
	//return SetAccessGrantedForAllUserChats(whc, whc.BotUserKey, value) // TODO: Call in deferrer
}

//func SetAccessGrantedForAllUserChats(whcb *WebhookContextBase, botUserKey *datastore.Key, value bool) error {
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
