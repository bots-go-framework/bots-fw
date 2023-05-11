package botsfw

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
)

// SetAccessGranted marks current context as authenticated
func SetAccessGranted(whc WebhookContext, value bool) (err error) {
	c := whc.Context()
	log.Debugf(c, "SetAccessGranted(value=%v)", value)
	botID := whc.GetBotCode()
	chatData := whc.ChatData()
	store := whc.Store()
	if chatData != nil {
		if chatData.IsAccessGranted() == value {
			log.Infof(c, "No need to change chatData.AccessGranted, as already is: %v", value)
		} else {
			chatKey := botsfwmodels.ChatKey{
				BotID: botID,
			}
			if chatKey.ChatID, err = whc.BotChatID(); err != nil {
				return err
			}
			if err = store.RunInTransaction(c, botID, func(c context.Context) error {
				if changed := chatData.SetAccessGranted(value); changed {
					if err = store.SaveBotChatData(c, chatKey, chatData); err != nil {
						err = fmt.Errorf("failed to save bot chat entity to db: %w", err)
					}
				}
				return nil
			}); err != nil {
				return
			}
		}
	}

	botUserID := whc.GetSender().GetID()
	botUserStrID := fmt.Sprintf("%v", botUserID)
	log.Debugf(c, "SetAccessGranted(): whc.GetSender().GetID() = %v", botUserID)
	if botUser, err := whc.Store().GetBotUserByID(c, botID, botUserStrID); err != nil {
		return fmt.Errorf("failed to get bot user by id=%v: %w", botUserID, err)
	} else if botUser.IsAccessGranted() == value {
		log.Infof(c, "No need to change botUser.AccessGranted, as already is: %v", value)
	} else if err = store.RunInTransaction(c, botID, func(c context.Context) error {
		botUser.SetAccessGranted(value)
		if botUser, err = whc.Store().GetBotUserByID(c, botID, botUserStrID); err != nil {
			return fmt.Errorf("failed to get transactionally bot user by id=%v: %w", botUserID, err)
		}
		if changed := botUser.SetAccessGranted(value); changed {
			if err = store.SaveBotUser(c, botID, botUserStrID, botUser); err != nil {
				err = fmt.Errorf("failed to call whc.SaveBotUser(botUserID=%v): %w", botUserID, err)
			}
		}
		return err
	}); err != nil {
		return err
	}
	return nil
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
