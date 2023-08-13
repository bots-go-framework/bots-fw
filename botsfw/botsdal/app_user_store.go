package botsdal

//import (
//	"context"
//	"github.com/bots-go-framework/bots-fw-store/botsfwdal"
//	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
//	"github.com/dal-go/dalgo/dal"
//	"reflect"
//)
//
//var _ botsfwdal.AppUserStore = (*appUserStore)(nil)
//
//type appUserStore struct {
//	//dalgoStore
//}
//
//func (store appUserStore) GetAppUserByID(c context.Context, botID, appUserID string, appUser botsfwmodels.AppUserData) error {
//	key := dal.NewKeyWithID(store.collection, appUserID)
//	record := dal.NewRecordWithData(key, appUser)
//	db, err := store.getDb(c, botID)
//	if err != nil {
//		return err
//	}
//	var getter dal.Getter = db
//	if tx, ok := dal.GetTransaction(c).(dal.ReadwriteTransaction); ok && tx != nil {
//		getter = tx
//	}
//
//	if err = getter.Get(c, record); err != nil {
//		if dal.IsNotFound(err) {
//			err = botsfwdal.NotFoundErr(err)
//		}
//		return err
//	}
//	return nil
//}
//
//func (store appUserStore) CreateAppUser(c context.Context, botID string, appUserData botsfwmodels.AppUserData) (appUserID string, err error) {
//	record := dal.NewRecordWithIncompleteKey(store.collection, reflect.String, appUserData)
//	return appUserID, store.runReadwriteTransaction(c, botID, func(c context.Context, tx dal.ReadwriteTransaction) error {
//		if err = tx.Insert(c, record); err != nil {
//			return err
//		}
//		appUserID = record.Key().ID.(string)
//		return nil
//	})
//}
//
//func (store appUserStore) SaveAppUser(c context.Context, botID, appUserID string, appUserData botsfwmodels.AppUserData) (err error) {
//	record := dal.NewRecordWithData(dal.NewKeyWithID(store.collection, appUserID), appUserData)
//	return store.runReadwriteTransaction(c, botID, func(c context.Context, tx dal.ReadwriteTransaction) error {
//		return tx.Set(c, record)
//	})
//}
//
//func newAppUserStore(collection string, getDb DbProvider) appUserStore {
//	if collection == "" {
//		panic("collection is empty")
//	}
//	if getDb == nil {
//		panic("getDb is nil")
//	}
//	return appUserStore{
//		dalgoStore: dalgoStore{
//			getDb:      getDb,
//			collection: collection,
//		},
//	}
//}
