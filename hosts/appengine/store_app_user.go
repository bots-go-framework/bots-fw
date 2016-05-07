package gae_host

import (
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
)

type GaeAppUserStore struct {
	appUserEntityKind string
	GaeBaseStore
}

var _ bots.AppUserStore = (*GaeAppUserStore)(nil)

func NewGaeAppUserStore(c context.Context, appUserEntityKind string) GaeAppUserStore {
	return GaeAppUserStore{
		GaeBaseStore: GaeBaseStore{
			c: c,
			entityKind: appUserEntityKind,
		},
	}
}

// ************************** Helper functions **************************

func (s GaeAppUserStore) appUserKey(appUserId int64) *datastore.Key {
	return datastore.NewKey(s.c, s.appUserEntityKind, "", appUserId, nil)
}

// ************************** Implementations of  bots.AppUserStore **************************
func (s GaeAppUserStore) GetAppUserByID(appUserId int64, appUser bots.AppUser) error {
	return nds.Get(s.c, s.appUserKey((int64)(appUserId)), appUser)
}

func (s GaeAppUserStore) CreateAppUser(appUserEntity bots.AppUser) (int64, error) {
	key, err := nds.Put(s.c, s.appUserKey(0), appUserEntity)
	return key.IntID(), err
}

func (s GaeAppUserStore) SaveAppUser(appUserId int64, appUserEntity bots.AppUser) error {
	if appUserId == 0 {
		panic("appUserId == 0")
	}
	_, err := nds.Put(s.c, s.appUserKey(appUserId), appUserEntity)
	return err
}
