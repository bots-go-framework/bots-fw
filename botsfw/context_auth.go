package botsfw

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfw/botsdal"
	"time"
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
					now := time.Now()
					chatDataBase := chatData.Base()
					chatDataBase.DtUpdated = now
					chatDataBase.SetDtLastInteraction(now) // Must set DtLastInteraction through wrapper
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
	tx := whc.Tx()
	platformID := whc.BotPlatform().ID()
	botSettings := whc.BotContext().BotSettings

	if botUser, err := botsdal.GetBotUser(c, tx, platformID, botSettings.Code, botUserStrID, botSettings.Profile.NewBotUserData); err != nil {
		return fmt.Errorf("failed to get bot user by id=%v: %w", botUserID, err)
	} else if botUser.Data.IsAccessGranted() == value {
		log.Infof(c, "No need to change botUser.AccessGranted, as already is: %v", value)
	} else if changed := botUser.Data.SetAccessGranted(value); changed {
		if err = tx.Set(c, botUser.Record); err != nil {
			err = fmt.Errorf("failed to save bot user record (key=%v): %w", botUser.Key, err)
		}
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
