package gae_host

import (
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"google.golang.org/appengine/datastore"
	"strconv"
	"time"
)

type EntityTypeValidator interface {
}

// Persist chat to GAE datastore
type GaeBotChatStore struct {
	GaeBaseStore
	botChats                  map[string]bots.BotChat
	NewBotChatKey             func(c context.Context, botID, botChatId string) *datastore.Key
	validateBotChatEntityType func(entity bots.BotChat)
	newBotChatEntity          func() bots.BotChat
}

var _ bots.BotChatStore = (*GaeBotChatStore)(nil) // Check for interface implementation at compile time

// ************************** Implementations of  bots.ChatStore **************************
func (s *GaeBotChatStore) GetBotChatEntityByID(c context.Context, botID, botChatID string) (botChatEntity bots.BotChat, err error) { // Former LoadBotChatEntity
	//log.Debugf(c, "GaeBotChatStore.GetBotChatEntityByID(%v)", botChatId)
	if s.botChats == nil {
		s.botChats = make(map[string]bots.BotChat, 1)
	}
	botChatEntity = s.newBotChatEntity()
	botChatKey := s.NewBotChatKey(c, botID, botChatID)
	//c, _ = context.WithDeadline(c, time.Now().Add(time.Second))
	if err = nds.Get(c, botChatKey, botChatEntity); err != nil {
		if err != datastore.ErrNoSuchEntity {
			return
		} else {
			if s.entityKind == "TgChat" { // TODO: Remove workaround to fix old entities
				var tgChatID int64
				if tgChatID, err = strconv.ParseInt(botChatID, 10, 64); err != nil {
					err = errors.Wrap(err, "Failed to parse botChatID to int")
					return
				} else {
					intKey := datastore.NewKey(c, s.entityKind, "", tgChatID, nil)
					if err = nds.Get(c, intKey, botChatEntity); err != nil {
						if err == datastore.ErrNoSuchEntity {
							log.Infof(c, errors.Wrapf(err, "There is no bot chat entity with intID=%v", intKey.IntID()).Error())
							err = bots.ErrEntityNotFound
						}
						return
					} else {
						log.Infof(c, "Telegram chat entity Found by int ID, will attempt to migrate...")
						if err = nds.RunInTransaction(c, func(tc context.Context) (err error) {
							if err = nds.Get(tc, intKey, botChatEntity); err != nil {
								return
							}
							if err = nds.Delete(tc, intKey); err != nil {
								return
							}
							if _, err = nds.Put(tc, botChatKey, botChatEntity); err != nil {
								return
							}
							return
						}, &datastore.TransactionOptions{XG: true}); err != nil {
							log.Errorf(c, errors.Wrap(err, "Failed to migrate Telegram chat entity").Error())
							return

						}
						log.Infof(c, "Telegram chat entity migrated to new key: [%v]", botChatKey.StringID())
					}
				}
			}
			err = bots.ErrEntityNotFound
			return
		}
	}
	if err == nil {
		s.botChats[botChatKey.StringID()] = botChatEntity
	}
	return
}

func (s *GaeBotChatStore) SaveBotChat(c context.Context, botID, botChatID string, chatEntity bots.BotChat) error { // Former SaveBotChatEntity
	s.validateBotChatEntityType(chatEntity)
	chatEntity.SetDtUpdated(time.Now())
	_, err := nds.Put(c, s.NewBotChatKey(c, botID, botChatID), chatEntity)
	return err
}

func (s *GaeBotChatStore) NewBotChatEntity(c context.Context, botID string, botChat bots.WebhookChat, appUserID int64, botUserID string, isAccessGranted bool) bots.BotChat {
	botChatId := botChat.GetID()
	log.Debugf(c, "NewBotChatEntity(botID=%v, botChatId=%v, appUserID=%v, botUserID=%v, isAccessGranted=%v)", botID, botChatId, appUserID, botUserID, isAccessGranted)
	botChatEntity := s.newBotChatEntity()
	botChatEntity.SetBotID(botID)

	if botChat.IsGroupChat() {
		botChatEntity.SetIsGroupChat(true)
	} else {
		botChatEntity.SetAppUserIntID(appUserID)
		botChatEntity.SetBotUserID(botUserID)
	}

	botChatEntity.SetAccessGranted(isAccessGranted)
	s.botChats[s.NewBotChatKey(c, botID, botChatId).StringID()] = botChatEntity // TODO: No need to create a key instance, create dedicated func to create ID?
	return botChatEntity
}

func (s *GaeBotChatStore) Close(c context.Context) error { // Former SaveBotChatEntity
	if len(s.botChats) == 0 {
		log.Debugf(c, "GaeBotChatStore.Close(): Nothing to save")
		return nil
	}
	//log.Debugf(c, "GaeBotChatStore.Close(): %v entities to save", len(s.botChats))
	var chatKeys []*datastore.Key
	var chatEntities []bots.BotChat
	now := time.Now()
	for chatId, chatEntity := range s.botChats {
		s.validateBotChatEntityType(chatEntity)
		chatEntity.SetDtUpdated(now)
		chatEntity.SetDtLastInteraction(now)
		chatKeys = append(chatKeys, datastore.NewKey(c, s.entityKind, chatId, 0, nil))
		chatEntities = append(chatEntities, chatEntity)
	}
	_, err := nds.PutMulti(c, chatKeys, chatEntities)
	if err == nil {
		//log.Debugf(c, "Successfully saved %v BotChat entities with keys: %v", len(chatKeys), chatKeys)
		s.botChats = nil
	} else {
		log.Errorf(c, "Failed to save %v BotChat entities: %v", len(chatKeys), err)
	}
	return err
}
