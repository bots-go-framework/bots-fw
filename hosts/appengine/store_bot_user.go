package gae_host

import (
	"github.com/qedus/nds"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/bots-framework/core"
)

// Persist user to GAE datastore
type GaeBotUserStore struct {
	GaeBaseStore
	//botUsers 					  map[interface{}]bots.BotUser
	botUserKey                func(botUserId interface{}) *datastore.Key
	validateBotUserEntityType func(entity bots.BotUser)
	newBotUserEntity          func() bots.BotUser
}
var _ bots.BotUserStore = (*GaeBotUserStore)(nil) // Check for interface implementation at compile time


// ************************** Implementations of  bots.BotUserStore **************************
func (s GaeBotUserStore) GetBotUserById(botUserId string) (bots.BotUser, error) { // Former LoadBotUserEntity
	//if s.botUsers == nil {
	//	s.botUsers = make(map[int]bots.BotUser, 1)
	//}
	botUserEntity := s.newBotUserEntity()
	err := nds.Get(s.c, s.botUserKey(botUserId), botUserEntity)
	return botUserEntity, err
}

func (s GaeBotUserStore) SaveBotUser(userId string, userEntity bots.BotUser) error { // Former SaveBotUserEntity
	s.validateBotUserEntityType(userEntity)
	_, err := nds.Put(s.c, s.botUserKey(userId), userEntity)
	return err
}
