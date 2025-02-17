package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
)

const botUsersCollection = "botUsers"

// NewPlatformUserKey creates a dalgo key to specific bot user record
func NewPlatformUserKey(platformID, botUserID string) *dal.Key {
	platformKey := NewPlatformKey(platformID)
	if botUserID == "" {
		panic("botUserID is required parameter")
	}
	return dal.NewKeyWithParentAndID(platformKey, botUsersCollection, botUserID)
}

// GetPlatformUser loads bot user data
func GetPlatformUser(
	ctx context.Context,
	tx dal.ReadSession,
	platformID, botUserID string,
	platformUserData botsfwmodels.PlatformUserData,
) (botUser BotUser, err error) {
	botUserKey := NewPlatformUserKey(platformID, botUserID)
	botUser = BotUser(record.NewDataWithID(botUserID, botUserKey, platformUserData))
	return botUser, tx.Get(ctx, botUser.Record)
}

// CreatePlatformUserRecord creates bot user record in database
func CreatePlatformUserRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	platformID, botUserID string,
	platformUserData botsfwmodels.PlatformUserData,
) (err error) {
	if validatableData, ok := platformUserData.(interface{ Validate() error }); ok {
		if err = validatableData.Validate(); err != nil {
			return err
		}
	}
	key := NewPlatformUserKey(platformID, botUserID)
	platformUser := record.NewDataWithID(botUserID, key, platformUserData)
	err = tx.Insert(ctx, platformUser.Record)
	return err
}

//import (
//	"context"
//	"fmt"
//	"github.com/bots-go-framework/bots-fw-store/botsfwdal"
//	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
//	"github.com/bots-go-framework/bots-fw/botsfw"
//	"github.com/dal-go/dalgo/dal"
//	"github.com/dal-go/dalgo/record"
//)
//
//var _ botsfwdal.BotUserStore = (*botUserStore)(nil)
//
//type botUserStore struct {
//	dalgoStore
//	platform       string
//	newBotUserData func(botID string) (botsfwmodels.PlatformUserData, error)
//	createBotUser  BotUserCreator
//}
//
//// newBotUserStore creates new bot user store
//func newBotUserStore(collection, platform string, getDb DbProvider, newBotUserData func(botID string) (botsfwmodels.PlatformUserData, error), createBotUser BotUserCreator) botUserStore {
//	if getDb == nil {
//		panic("getDb is nil")
//	}
//	if collection == "" {
//		panic("collection is empty")
//	}
//	if newBotUserData == nil {
//		panic("newBotUserData is nil")
//	}
//	return botUserStore{
//		dalgoStore: dalgoStore{
//			getDb:      getDb,
//			collection: collection,
//		},
//		platform:       platform,
//		newBotUserData: newBotUserData,
//		createBotUser:  createBotUser,
//	}
//}
//
//type botUserWithStrID struct {
//	record.WithID[string]
//	Data botsfwmodels.PlatformUserData
//}
//
//// GetBotUserByID returns bot user data
//func (store botUserStore) GetBotUserByID(c context.Context, botID, botUserID string) (botUserData botsfwmodels.PlatformUserData, err error) {
//	key := store.botUserRecordKey(botUserID)
//	if botUserData, err = store.newBotUserData(botID); err != nil {
//		return
//	}
//	botUser := botUserWithStrID{
//		Data: botUserData,
//		WithID: record.WithID[string]{
//			ID:     botUserID,
//			Record: dal.NewRecordWithData(key, botUserData),
//		},
//	}
//	var db dal.Database
//	if db, err = store.getDb(c, botID); err != nil {
//		err = fmt.Errorf("failed to get getDb: %w", err)
//		return
//	}
//
//	var getter dal.Getter = db
//	if tx, ok := dal.GetTransaction(c).(dal.ReadwriteTransaction); ok && tx != nil {
//		getter = tx
//	}
//
//	if err = getter.Get(c, botUser.Record); err != nil {
//		if dal.IsNotFound(err) {
//			err = botsfwdal.NotFoundErr(err)
//		}
//		return
//	}
//	return
//}
//
//// SaveBotUser saves bot user data
//func (store botUserStore) SaveBotUser(c context.Context, botID, botUserID string, botUserData botsfwmodels.PlatformUserData) error {
//	key := store.botUserRecordKey(botUserID)
//	botUserRecord := dal.NewRecordWithData(key, botUserData)
//	return store.runReadwriteTransaction(c, botID, func(c context.Context, tx dal.ReadwriteTransaction) error {
//		return tx.Set(c, botUserRecord)
//	})
//}
//
//func (store botUserStore) CreatePlatformUserRecord(c context.Context, botID string, apiUser botsfw.WebhookActor) (botsfwmodels.PlatformUserData, error) {
//	return store.createBotUser(c, botID, apiUser)
//}
//
//func (store botUserStore) botUserRecordKey(botUserID any) *dal.Key {
//	id := fmt.Sprintf("%s:%s", store.platform, botUserID)
//	return dal.NewKeyWithID(store.collection, id)
//}
