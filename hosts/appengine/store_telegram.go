package gae_host

import (
	"github.com/strongo/bots-framework/core"
	//"github.com/strongo/bots-framework/platforms/telegram"
	//"google.golang.org/appengine/datastore"
	//"time"
	//"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	//"github.com/qedus/nds"
	"golang.org/x/net/context"
	"github.com/strongo/bots-api-telegram"
)

func GetOrCreateUserEntity(ctx context.Context, update tgbotapi.Update) (bots.BotUser, error) {
	return nil, bots.NotImplementedError
	//from := update.Message.From
	//var telegramUser telegram_bot.TelegramUser
	//telegramUser := new(telegram_bot.TelegramUser)
	//err := bots.LoadBotUserEntity(ctx, whc.UserKey(), &telegramUser)
	//if err == datastore.ErrNoSuchEntity {
	//	telegramUser.DtCreated = time.Now()
	//	telegramUser.FirstName = from.FirstName
	//	telegramUser.LastName = from.LastName
	//	telegramUser.UserName = from.UserName
	//
	//	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
	//		userKey := datastore.NewIncompleteKey(ctx, common.AppUserKind, nil)
	//		user := common.AppUser{
	//			DtCreated:      telegramUser.DtCreated,
	//			IsTelegramUser: true,
	//			//TelegramUserIDs: []int64{int64(from.ID)},
	//			FirstName: from.FirstName,
	//			LastName:  from.LastName,
	//			UserName:  from.UserName,
	//		}
	//		userKey, err := nds.Put(ctx, userKey, &user)
	//		if err != nil {
	//			log.Errorf(ctx, "Failed to create new User: %v", err)
	//			return err
	//		}
	//		telegramUser.AppUserID = userKey.IntID()
	//		_, err = nds.Put(ctx, whc.TelegramUserKey(ctx), telegramUser)
	//		if err != nil {
	//			log.Errorf(ctx, "Failed to create new TelegramUser: %v", err)
	//			return err
	//		}
	//		return err
	//	}, &datastore.TransactionOptions{XG: true})
	//	if err != nil {
	//		log.Errorf(ctx, "Failed to create new User & TelegramUser entities")
	//	} else {
	//		log.Infof(ctx, "Created new User & TelegramUser entities")
	//	}
	//}
	//return telegramUser, err
}
