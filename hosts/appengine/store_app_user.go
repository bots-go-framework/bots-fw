package gae_host

import (
	"github.com/qedus/nds"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"net/http"
)

type GaeAppUserStore struct {
	appUserEntityKind string
	newUserEntity     func() bots.AppUser
	GaeBaseStore
}

var _ bots.AppUserStore = (*GaeAppUserStore)(nil)

func NewGaeAppUserStore(log bots.Logger, r *http.Request, appUserEntityKind string, newUserEntity func() bots.AppUser) GaeAppUserStore {
	return GaeAppUserStore{
		appUserEntityKind: appUserEntityKind,
		newUserEntity:     newUserEntity,
		GaeBaseStore:      NewGaeBaseStore(log, r, appUserEntityKind),
	}
}

// ************************** Helper functions **************************

func (s GaeAppUserStore) appUserKey(appUserId int64) *datastore.Key {
	return datastore.NewKey(s.Context(), s.appUserEntityKind, "", appUserId, nil)
}

// ************************** Implementations of  bots.AppUserStore **************************
func (s GaeAppUserStore) GetAppUserByID(appUserId int64, appUser bots.AppUser) error {
	return nds.Get(s.Context(), s.appUserKey(appUserId), appUser)
}

func (s GaeAppUserStore) CreateAppUser(actor bots.WebhookActor) (int64, bots.AppUser, error) {
	return s.createAppUser(s.Context(), actor)
}

func (s GaeAppUserStore) createAppUser(c context.Context, actor bots.WebhookActor) (int64, bots.AppUser, error) {
	appUserEntity := s.newUserEntity()
	appUserEntity.SetNames(actor.GetFirstName(), actor.GetLastName(), actor.GetUserName())
	key, err := nds.Put(c, s.appUserKey(0), appUserEntity)
	return key.IntID(), appUserEntity, err
}

func (s GaeAppUserStore) SaveAppUser(appUserId int64, appUserEntity bots.AppUser) error {
	if appUserId == 0 {
		panic("appUserId == 0")
	}
	_, err := nds.Put(s.Context(), s.appUserKey(appUserId), appUserEntity)
	return err
}
