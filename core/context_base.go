package bots

import (
	"google.golang.org/appengine/datastore"
	"net/http"
	"google.golang.org/appengine"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type WebhookContextBase struct {
	w http.ResponseWriter
	r *http.Request

	BotHost

	locale      Locale
	BotSettings BotSettings
	//update      tgbotapi.Update
	chatKey     *datastore.Key
	chatEntity  BotChat

	BotUserKey     *datastore.Key
	user UserEntity
	Translator
	Locales LocalesProvider
}

func (whcb *WebhookContextBase) GetLogger() Logger{
	return whcb.BotHost.GetLogger(whcb.r)
}

func (whcb *WebhookContextBase) Translate(key string) string {
	return whcb.Translator.Translate(key, whcb.locale.Code5)
}

func (whcb *WebhookContextBase) TranslateNoWarning(key string) string {
	return whcb.Translator.TranslateNoWarning(key, whcb.locale.Code5)
}

func (whcb *WebhookContextBase) GetHttpClient() *http.Client {
	return whcb.BotHost.GetHttpClient(whcb.r)
}

func (whcb *WebhookContextBase) HasChatEntity() bool {
	return whcb.chatEntity != nil
}

func (whcb *WebhookContextBase) SetChatEntity(chatEntity BotChat) {
	whcb.chatEntity = chatEntity
}

func (whcb *WebhookContextBase) ChatEntity(whc WebhookContext) BotChat {
	if whcb.chatEntity == nil {
		chatEntity, _ := whcb.GetChatEntity(whc)
		whcb.SetChatEntity(chatEntity)
	}
	return whcb.chatEntity
}

func (whcb *WebhookContextBase) GetChatEntity(whc WebhookContext) (BotChat, error) {
	ctx := whc.Context()
	if whcb.HasChatEntity() {
		log.Warningf(ctx, "Duplicate call of func (whc *bot.WebhookContext) _getChat()")
		return whcb.chatEntity, nil
	}

	chatEntity := whc.NewChatEntity()

	err := LoadBotChatEntity(ctx, whc.ChatKey(), chatEntity)
	switch err {
	case nil: // Nothing to do
	case datastore.ErrNoSuchEntity: //TODO: Should be this moved to DAL?
		err = nil
		log.Infof(ctx, "Creating new BotChat entity...")
		chatEntity = whc.MakeChatEntity()
		userEntity, err := whc.GetOrCreateUserEntity()
		if err == nil {
			chatEntity.SetUserID(userEntity.GetUserID())
			if userEntity.IsAccessGranted() {
				chatEntity.SetAccessGranted(true)
			}
		}
	default:
		log.Errorf(ctx, "Failed to load TelegramChat: %v", err)
		return nil, err
	}
	log.Debugf(ctx, `chatEntity.PreferredLanguage: %v, whc.locale.Code5: %v, chatEntity.PreferredLanguage != """ && whc.locale.Code5 != chatEntity.PreferredLanguage: %v`, chatEntity.GetPreferredLanguage(), whc.Locale().Code5, chatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != chatEntity.GetPreferredLanguage())
	if chatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != chatEntity.GetPreferredLanguage() {
		err = whc.SetLocale(chatEntity.GetPreferredLanguage())
		if err != nil {

		} else {
			log.Debugf(ctx, "whc.locale cahged to: %v", whc.Locale)
		}
	}
	return chatEntity, err
}
func NewWebhookContextBase(botHost BotHost, w http.ResponseWriter, r *http.Request, translator Translator) *WebhookContextBase {
	return &WebhookContextBase{w: w, r: r, BotHost: botHost, Translator: translator}
}

func (whc *WebhookContextBase) ChatKey() *datastore.Key {
	return whc.chatKey
}

func (whc *WebhookContextBase) NewChatKey(c context.Context) *datastore.Key {
	chatKey := whc.ChatKey()
	return datastore.NewKey(c, chatKey.Kind(), chatKey.StringID(), chatKey.IntID(), chatKey.Parent())
}

func (whc *WebhookContextBase) SetChatKey(key *datastore.Key)  {
	whc.chatKey = key
}

func (whc *WebhookContextBase) UserEntity() UserEntity {
	return whc.user
}


func (c *WebhookContextBase) InitBase(r *http.Request, botSettings BotSettings) {
	c.r = r
	c.locale = botSettings.Locale
	c.BotSettings = botSettings
}

func (whc *WebhookContextBase) Context() context.Context {
	return appengine.NewContext(whc.r)
}

func (whcb *WebhookContextBase) NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot {
	return MessageFromBot{Text: fmt.Sprintf(whcb.Translate(messageCode), a...)}
}

func (*WebhookContextBase) NewMessage(text string) MessageFromBot {
	return MessageFromBot{Text: text}
}

func (whcb *WebhookContextBase) Locale() Locale {
	return whcb.locale
}

func (whcb *WebhookContextBase) SetLocale(code5 string) error {
	locale, err := whcb.Locales.GetLocaleByCode5(code5)
	if err == nil {
		whcb.locale = locale
	}
	return err
}
