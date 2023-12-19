package botsdal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
)

const botUsersCollection = "botUsers"

// NewBotUserKey creates a dalgo key to specific bot user record
func NewBotUserKey(platformID, botID, botUserID string) *dal.Key {
	botKey := NewBotKey(platformID, botID)
	if botUserID == "" {
		panic("botUserID is required parameter")
	}
	return dal.NewKeyWithParentAndID(botKey, botUsersCollection, botUserID)
}

// GetBotUser loads bot user data
func GetBotUser(
	ctx context.Context,
	tx dal.ReadSession,
	platformID, botID, botUserID string,
	newData func() botsfwmodels.BotUserData,
) (botUser record.DataWithID[string, botsfwmodels.BotUserData], err error) {
	botUserKey := NewBotUserKey(platformID, botID, botUserID)
	data := newData()
	botUser = record.NewDataWithID(botUserID, botUserKey, data)
	return botUser, tx.Get(ctx, botUser.Record)
}

// CreateBotUserRecord creates bot user record in database
func CreateBotUserRecord(
	ctx context.Context,
	tx dal.ReadwriteTransaction,
	platformID, botID, botUserID string,
	botUserData botsfwmodels.BotUserData,
) (err error) {
	if validatableData, ok := botUserData.(interface{ Validate() error }); ok {
		if err = validatableData.Validate(); err != nil {
			return err
		}
	}
	botUserKey := NewBotUserKey(platformID, botID, botUserID)
	botUser := record.NewDataWithID(botUserID, botUserKey, botUserData)
	err = tx.Insert(ctx, botUser.Record)
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
//	newBotUserData func(botID string) (botsfwmodels.BotUserData, error)
//	createBotUser  BotUserCreator
//}
//
//// newBotUserStore creates new bot user store
//func newBotUserStore(collection, platform string, getDb DbProvider, newBotUserData func(botID string) (botsfwmodels.BotUserData, error), createBotUser BotUserCreator) botUserStore {
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
//	Data botsfwmodels.BotUserData
//}
//
//// GetBotUserByID returns bot user data
//func (store botUserStore) GetBotUserByID(c context.Context, botID, botUserID string) (botUserData botsfwmodels.BotUserData, err error) {
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
//func (store botUserStore) SaveBotUser(c context.Context, botID, botUserID string, botUserData botsfwmodels.BotUserData) error {
//	key := store.botUserRecordKey(botUserID)
//	botUserRecord := dal.NewRecordWithData(key, botUserData)
//	return store.runReadwriteTransaction(c, botID, func(c context.Context, tx dal.ReadwriteTransaction) error {
//		return tx.Set(c, botUserRecord)
//	})
//}
//
//func (store botUserStore) CreateBotUserRecord(c context.Context, botID string, apiUser botsfw.WebhookActor) (botsfwmodels.BotUserData, error) {
//	return store.createBotUser(c, botID, apiUser)
//}
//
//func (store botUserStore) botUserRecordKey(botUserID any) *dal.Key {
//	id := fmt.Sprintf("%s:%s", store.platform, botUserID)
//	return dal.NewKeyWithID(store.collection, id)
//}
