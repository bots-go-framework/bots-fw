package gae_host

//import (
//	"github.com/strongo/bots-framework/core"
//	//"github.com/strongo/bots-framework/platforms/telegram"
//	//"google.golang.org/appengine/datastore"
//	//"time"
//	//"bitbucket.com/debtstracker/gae_app/debtstracker/common"
//	//"github.com/qedus/nds"
//	"golang.org/x/net/context"
//	"github.com/strongo/bots-api-telegram"
//	"github.com/strongo/bots-framework/platforms/telegram"
//	"google.golang.org/appengine/datastore"
//	"time"
//	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
//	"github.com/qedus/nds"
//)

//func GetOrCreateUserEntity(log bots.Logger, ctx context.Context, sender bots.WebhookSender) (bots.BotUser, error) {
//	var telegramUser telegram_bot.TelegramUser
//	telegramUser := new(telegram_bot.TelegramUser)
//	err := bots.LoadBotUserEntity(ctx, whc.UserKey(), &telegramUser)
//	if err == datastore.ErrNoSuchEntity {
//		firstName := sender.GetFirstName()
//		lastName := sender.GetLastName()
//		userName := sender.GetUserName()
//
//		telegramUser.DtCreated = time.Now()
//		telegramUser.FirstName = firstName
//		telegramUser.LastName = lastName
//		telegramUser.UserName = userName
//
//		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
//			userKey := datastore.NewIncompleteKey(ctx, common.AppUserKind, nil)
//			user := common.AppUser{
//				DtCreated:      telegramUser.DtCreated,
//				IsTelegramUser: true,
//				//TelegramUserIDs: []int64{int64(from.ID)},
//				FirstName: firstName,
//				LastName:  lastName,
//				UserName:  userName,
//			}
//			userKey, err := nds.Put(ctx, userKey, &user)
//			if err != nil {
//				log.Errorf(ctx, "Failed to create new User: %v", err)
//				return err
//			}
//			telegramUser.AppUserIntID = userKey.IntID()
//			_, err = nds.Put(ctx, whc.TelegramUserKey(ctx), telegramUser)
//			if err != nil {
//				log.Errorf(ctx, "Failed to create new TelegramUser: %v", err)
//				return err
//			}
//			return err
//		}, &datastore.TransactionOptions{XG: true})
//		if err != nil {
//			log.Errorf(ctx, "Failed to create new User & TelegramUser entities")
//		} else {
//			log.Infof(ctx, "Created new User & TelegramUser entities")
//		}
//	}
//	return telegramUser, err
//}
