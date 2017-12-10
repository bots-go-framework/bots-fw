package gae_host

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

// Persist user to GAE datastore
type GaeBotUserStore struct {
	GaeBaseStore
	//botUsers 					  map[interface{}]bots.BotUser
	botUserKey                func(c context.Context, botUserID interface{}) *datastore.Key
	validateBotUserEntityType func(entity bots.BotUser)
	newBotUserEntity          func(apiUser bots.WebhookActor) bots.BotUser
	gaeAppUserStore           GaeAppUserStore
}

var _ bots.BotUserStore = (*GaeBotUserStore)(nil) // Check for interface implementation at compile time

// ************************** Implementations of  bots.BotUserStore **************************
func (s GaeBotUserStore) GetBotUserById(c context.Context, botUserId interface{}) (bots.BotUser, error) { // Former LoadBotUserEntity
	//if s.botUsers == nil {
	//	s.botUsers = make(map[int]bots.BotUser, 1)
	//}
	botUserEntity := s.newBotUserEntity(nil)
	err := nds.Get(c, s.botUserKey(c, botUserId), botUserEntity)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil
	}
	return botUserEntity, err
}

func (s GaeBotUserStore) SaveBotUser(c context.Context, botUserID interface{}, userEntity bots.BotUser) error { // Former SaveBotUserEntity
	// TODO: Architecture needs refactoring as it not transactional save
	// We load bot user entity outside of here (out of transaction) and save here. It can change since then.
	s.validateBotUserEntityType(userEntity)
	userEntity.SetDtUpdated(time.Now())
	err := nds.RunInTransaction(c, func(c context.Context) error {
		key := s.botUserKey(c, botUserID)
		existingBotUser := s.newBotUserEntity(nil)
		err := nds.Get(c, key, existingBotUser)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				err = nil
			}
		} else {
			if existingBotUser.GetAppUserIntID() != userEntity.GetAppUserIntID() {
				return fmt.Errorf(
					"Data integrity issue, existingBotUser.GetAppUserIntID():%v != userEntity.GetAppUserIntID():%v",
					existingBotUser.GetAppUserIntID(),
					userEntity.GetAppUserIntID(),
				)
			}
		}
		_, err = nds.Put(c, key, userEntity)
		if err != nil {
			err = errors.Wrap(err, "SaveBotUser(): Failed to put user entity to datastore")
		}
		return err
	}, nil)
	return err
}

func (s GaeBotUserStore) CreateBotUser(c context.Context, botID string, apiUser bots.WebhookActor) (bots.BotUser, error) {
	log.Debugf(c, "GaeBotUserStore.CreateBotUser(botID=%v, apiUser=%T) started...", botID, apiUser)
	botUserID := apiUser.GetID()
	botUserEntity := s.newBotUserEntity(apiUser)

	var (
		appUserId int64
		appUser   bots.BotAppUser
		newUser   bool
	)

	err := nds.RunInTransaction(c, func(ctx context.Context) (err error) {
		botUserKey := s.botUserKey(ctx, botUserID)
		err = nds.Get(ctx, botUserKey, botUserEntity)

		if err == datastore.ErrNoSuchEntity {
			appUserId, err := s.gaeAppUserStore.getAppUserIdByBotUserKey(c, botUserKey)
			if err != nil {
				return err
			}
			if appUserId == 0 {
				appUserId, appUser, err = s.gaeAppUserStore.createAppUser(ctx, botID, apiUser)
				if err != nil {
					log.Errorf(c, "Failed to create app user: %v", err)
					return err
				}
				newUser = true
			}
			botUserEntity.SetAppUserIntID(appUserId)
			botUserEntity.SetDtUpdated(time.Now())
			botUserKey, err = nds.Put(ctx, botUserKey, botUserEntity)
		} else if err != nil {
			return err
		}

		return nil
	}, &datastore.TransactionOptions{XG: true})

	if err != nil {
		return nil, err
	}

	if newUser && appUserId != 0 && appUser != nil {
		// Workaround - check for missing entity
		appUserKey := datastore.NewKey(c, "User", "", appUserId, nil)
		if err = nds.Get(c, appUserKey, botUserEntity); err != nil {
			if err == datastore.ErrNoSuchEntity {
				if err = nds.RunInTransaction(c, func(tc context.Context) (err error) {
					if err = nds.Get(c, appUserKey, make(datastore.PropertyList, 0)); err != nil {
						if err == datastore.ErrNoSuchEntity {
							_, err = nds.Put(c, appUserKey, appUser) // Try to re-create
						}
						log.Errorf(c, err.Error())
						err = nil
					}
					return
				}, nil); err != nil {
					return botUserEntity, err
				}
			}
			return botUserEntity, err
		}
	}

	return botUserEntity, nil
}
