package gae_host

import (
	"github.com/qedus/nds"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/bots-framework/core"
)

type EntityTypeValidator interface {

}

// Persist chat to GAE datastore
type GaeBotChatStore struct {
	GaeBaseStore
	botChats                  map[interface{}]bots.BotChat
	botChatKey                func(botChatId interface{}) *datastore.Key
	validateBotChatEntityType func(entity bots.BotChat)
	newBotChatEntity          func() bots.BotChat
}
var _ bots.BotChatStore = (*GaeBotChatStore)(nil) // Check for interface implementation at compile time

// ************************** Implementations of  bots.ChatStore **************************
func (s *GaeBotChatStore) GetBotChatEntityById(botChatId interface{}) (bots.BotChat, error) { // Former LoadBotChatEntity
	if s.botChats == nil {
		s.botChats = make(map[interface{}]bots.BotChat, 1)
	}
	botChatEntity := s.newBotChatEntity()
	err := nds.Get(s.Context(), s.botChatKey(botChatId), botChatEntity)
	if err == datastore.ErrNoSuchEntity {
		return nil, bots.ErrEntityNotFound
	}
	if err == nil {
		s.botChats[botChatId] = botChatEntity
	}
	return botChatEntity, err
}

func (s *GaeBotChatStore) SaveBotChat(chatId interface{}, chatEntity bots.BotChat) error { // Former SaveBotChatEntity
	s.validateBotChatEntityType(chatEntity)
	_, err := nds.Put(s.Context(), s.botChatKey(chatId), chatEntity)
	return err
}

func (s *GaeBotChatStore) CreateBotChat(appUserID int64, botUserID interface{}, isAccessGranted bool) bots.BotChat {
	botChat := s.newBotChatEntity()
	botChat.SetAppUserID(appUserID)
	botChat.SetBotUserID(botUserID)
	botChat.SetAccessGranted(isAccessGranted)
	return botChat
}

func (s *GaeBotChatStore) Close() error { // Former SaveBotChatEntity
	var chatKeys []*datastore.Key
	for chatId, chatEntity := range s.botChats {
		s.validateBotChatEntityType(chatEntity)
		chatEntity.SetDtUpdatedToNow()
		chatKeys = append(chatKeys, s.botChatKey(chatId))
	}
	_, err := nds.PutMulti(s.Context(), chatKeys, s.botChats)
	return err
}
