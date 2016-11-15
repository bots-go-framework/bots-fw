package gae_host

import (
	"github.com/qedus/nds"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"strconv"
	"github.com/pkg/errors"
)

type EntityTypeValidator interface {
}

// Persist chat to GAE datastore
type GaeBotChatStore struct {
	GaeBaseStore
	botChats                  map[string]bots.BotChat
	botChatKey                func(c context.Context, botID, botChatId string) *datastore.Key
	validateBotChatEntityType func(entity bots.BotChat)
	newBotChatEntity          func() bots.BotChat
}

var _ bots.BotChatStore = (*GaeBotChatStore)(nil) // Check for interface implementation at compile time

// ************************** Implementations of  bots.ChatStore **************************
func (s *GaeBotChatStore) GetBotChatEntityById(c context.Context, botID, botChatID string) (bots.BotChat, error) { // Former LoadBotChatEntity
	//s.logger.Debugf(c, "GaeBotChatStore.GetBotChatEntityById(%v)", botChatId)
	if s.botChats == nil {
		s.botChats = make(map[string]bots.BotChat, 1)
	}
	botChatEntity := s.newBotChatEntity()
	botChatKey := s.botChatKey(c, botID, botChatID)
	err := nds.Get(c, botChatKey, botChatEntity)
	if err != nil {
		s.logger.Infof(c, "Failed to get bot chat entity by ID: %v - %T(%v)", botChatID, err, err)
		if err == datastore.ErrNoSuchEntity {
			if s.entityKind == "TgChat" { // TODO: Remove workaround to fix old entities
				var tgChatID int64
				if tgChatID, err = strconv.ParseInt(botChatID, 10, 64); err != nil {
					return nil, errors.Wrap(err, "Failet to parse botChatID to int")
				} else {
					intKey := datastore.NewKey(c, s.entityKind, "", tgChatID, nil)
					if err = nds.Get(c, intKey, botChatEntity); err != nil {
						if err == datastore.ErrNoSuchEntity {
							s.logger.Infof(c, errors.Wrapf(err, "Failed to get bot chat entity by int ID=%v", intKey.IntID()).Error())
							return nil, bots.ErrEntityNotFound
						}
					} else {
						s.logger.Infof(c, "Telegram chat entity Found by int ID, will attempt to migrate...")
						err = nds.RunInTransaction(c, func(c context.Context) error {
							if err = nds.Get(c, intKey, botChatEntity); err == nil {
								if err = nds.Delete(c, intKey); err != nil {
									return err
								}
								if _, err = nds.Put(c, botChatKey, botChatEntity); err != nil {
									return err
								}
							}
							return err
						}, &datastore.TransactionOptions{XG: true})
						if err == nil {
							s.logger.Infof(c, "Telegram chat entity migrated to new key: [%v]", botChatKey.StringID())
						} else {
							s.logger.Errorf(c, errors.Wrap(err, "Failed to migrate Telegram chat entity").Error())
						}
					}
				}
			} else {
				return nil, bots.ErrEntityNotFound
			}
		}
	}
	if err == nil {
		s.botChats[botChatKey.StringID()] = botChatEntity
	}
	return botChatEntity, err
}

func (s *GaeBotChatStore) SaveBotChat(c context.Context, botID, botChatID string, chatEntity bots.BotChat) error { // Former SaveBotChatEntity
	s.validateBotChatEntityType(chatEntity)
	chatEntity.SetDtUpdatedToNow()
	_, err := nds.Put(c, s.botChatKey(c, botID, botChatID), chatEntity)
	return err
}

func (s *GaeBotChatStore) NewBotChatEntity(c context.Context, botID string, botChatId string, appUserID int64, botUserID string, isAccessGranted bool) bots.BotChat {
	s.logger.Debugf(c, "NewBotChatEntity(botID=%v, botChatId=%v, appUserID=%v, botUserID=%v, isAccessGranted=%v)", botID, botChatId, appUserID, botUserID, isAccessGranted)
	botChat := s.newBotChatEntity()
	botChat.SetAppUserIntID(appUserID)
	botChat.SetBotUserID(botUserID)
	botChat.SetAccessGranted(isAccessGranted)
	botChat.SetBotID(botID)
	s.botChats[s.botChatKey(c, botID, botChatId).StringID()] = botChat // TODO: No need to create a key instance, create dedicated func to create ID?
	return botChat
}

func (s *GaeBotChatStore) Close(c context.Context) error { // Former SaveBotChatEntity
	if len(s.botChats) == 0 {
		s.logger.Debugf(c, "GaeBotChatStore.Close(): Nothing to save")
		return nil
	}
	//s.logger.Debugf(c, "GaeBotChatStore.Close(): %v entities to save", len(s.botChats))
	var chatKeys []*datastore.Key
	var chatEntities []bots.BotChat
	for chatId, chatEntity := range s.botChats {
		s.validateBotChatEntityType(chatEntity)
		chatEntity.SetDtUpdatedToNow()
		chatEntity.SetDtLastInteractionToNow()
		chatKeys = append(chatKeys, datastore.NewKey(c, s.entityKind, chatId, 0, nil))
		chatEntities = append(chatEntities, chatEntity)
	}
	_, err := nds.PutMulti(c, chatKeys, chatEntities)
	if err == nil {
		s.logger.Infof(c, "Succesfully saved %v BotChat entities with keys: %v", len(chatKeys), chatKeys)
		s.botChats = nil
	} else {
		s.logger.Errorf(c, "Failed to save %v BotChat entities: %v", len(chatKeys), err)
	}
	return err
}
