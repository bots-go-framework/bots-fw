package bots

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
)

type WebhookContextBase struct {
	//w          http.ResponseWriter
	r          *http.Request

	BotContext BotContext
	WebhookInput

	locale     Locale

	//update      tgbotapi.Update
	chatEntity BotChat

	BotUserKey *datastore.Key
	appUser    AppUser
	Translator
	Locales    LocalesProvider

	BotChatStore
}

func NewWebhookContextBase(r *http.Request, botContext BotContext, webhookInput WebhookInput, botChatStore BotChatStore) *WebhookContextBase {
	return &WebhookContextBase{
		r: r,
		BotContext: botContext,
		WebhookInput: webhookInput,
		BotChatStore: botChatStore,
	}
}

func (whcb *WebhookContextBase) GetLogger() Logger {
	return whcb.BotContext.BotHost.GetLogger(whcb.r)
}

func (whcb *WebhookContextBase) Translate(key string) string {
	return whcb.Translator.Translate(key, whcb.locale.Code5)
}

func (whcb *WebhookContextBase) TranslateNoWarning(key string) string {
	return whcb.Translator.TranslateNoWarning(key, whcb.locale.Code5)
}

func (whcb *WebhookContextBase) GetHttpClient() *http.Client {
	return whcb.BotContext.BotHost.GetHttpClient(whcb.r)
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

	botChatID := whc.BotChatID()
	whc.GetLogger().Infof("botChatID: %v", botChatID)
	botChatEntity, err := whcb.BotChatStore.GetBotChatEntityById(botChatID)
	switch err {
	case nil: // Nothing to do
	case ErrEntityNotFound: //TODO: Should be this moved to DAL?
		err = nil
		log.Infof(ctx, "Creating new BotChat entity...")
		userEntity, err := whc.GetOrCreateUserEntity()
		if err == nil {
			botChatEntity.SetAppUserID(userEntity.GetUserID())
			if userEntity.IsAccessGranted() {
				botChatEntity.SetAccessGranted(true)
			}
		}
	default:
		log.Errorf(ctx, "Failed to load TelegramChat: %v", err)
		return nil, err
	}
	log.Debugf(ctx, `chatEntity.PreferredLanguage: %v, whc.locale.Code5: %v, chatEntity.PreferredLanguage != """ && whc.locale.Code5 != chatEntity.PreferredLanguage: %v`, botChatEntity.GetPreferredLanguage(), whc.Locale().Code5, botChatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != botChatEntity.GetPreferredLanguage())
	if botChatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != botChatEntity.GetPreferredLanguage() {
		err = whc.SetLocale(botChatEntity.GetPreferredLanguage())
		if err != nil {

		} else {
			log.Debugf(ctx, "whc.locale cahged to: %v", whc.Locale)
		}
	}
	return botChatEntity, err
}

func (whc *WebhookContextBase) AppUserEntity() AppUser {
	return whc.appUser
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
