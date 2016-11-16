package gae_host

import (
	"github.com/pkg/errors"
	"github.com/qedus/nds"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"fmt"
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
	userEntity.SetDtUpdatedToNow()
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
				return errors.New(fmt.Sprintf(
					"Data integrity issue, existingBotUser.GetAppUserIntID():%v != userEntity.GetAppUserIntID():%v",
					existingBotUser.GetAppUserIntID(),
					userEntity.GetAppUserIntID(),
				))
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
	s.logger.Debugf(c, "CreateBotUser() started...")
	botUserID := apiUser.GetID()
	botUserEntity := s.newBotUserEntity(apiUser)

	err := nds.RunInTransaction(c, func(ctx context.Context) error {
		botUserKey := s.botUserKey(ctx, botUserID)
		err := nds.Get(ctx, botUserKey, botUserEntity)

		if err == datastore.ErrNoSuchEntity {
			appUserId, err := s.gaeAppUserStore.getAppUserIdByBotUserKey(c, botUserKey)
			if err != nil {
				return err
			}
			if appUserId == 0 {
				appUserId, _, err = s.gaeAppUserStore.createAppUser(ctx, botID, 	apiUser)
				if err != nil {
					s.logger.Errorf(c, "Failed to create app user: %v", err)
					return err
				}
			}
			botUserEntity.SetAppUserIntID(appUserId)
			botUserEntity.SetDtUpdatedToNow()
			botUserKey, err = nds.Put(ctx, botUserKey, botUserEntity)
		} else if err != nil {
			return err
		}

		return nil
	}, &datastore.TransactionOptions{XG: true})

	if err != nil {
		return nil, err
	}
	return botUserEntity, nil
}
