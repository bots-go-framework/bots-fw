package bots

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
)

type WebhookContextBase struct {
	//w          http.ResponseWriter
	r          *http.Request

	AppContext AppContext
	BotContext BotContext
	WebhookInput

	locale     Locale

	//update      tgbotapi.Update
	chatEntity BotChat

	BotUserKey *datastore.Key
	appUser    AppUser
	Translator
	//Locales    LocalesProvider

	BotCoreStores
}

func NewWebhookContextBase(r *http.Request, appContext AppContext, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores) *WebhookContextBase {
	whcb := WebhookContextBase{
		r: r,
		AppContext: appContext,
		BotContext: botContext,
		WebhookInput: webhookInput,
		BotCoreStores: botCoreStores,
	}
	whcb.Translator = appContext.GetTranslator(whcb.GetLogger())
	return &whcb
}

func (whcb *WebhookContextBase) GetLogger() Logger {
	return whcb.BotContext.BotHost.GetLogger(whcb.r)
}

func (whcb *WebhookContextBase) GetBotCode() string {
	return whcb.BotContext.BotSettings.Code
}

func (whcb *WebhookContextBase) GetBotToken() string {
	return whcb.BotContext.BotSettings.Token
}

func (whcb *WebhookContextBase) Translate(key string) string {
	return whcb.Translator.Translate(key, whcb.Locale().Code5)
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

func (whcb *WebhookContextBase) GetAppUser() (AppUser, error) {
	appUserID := whcb.chatEntity.GetAppUserID()
	appUser := whcb.AppContext.NewAppUserEntity()
	err := whcb.AppUserStore.GetAppUserByID(appUserID, appUser)
	return appUser, err
}

func (whcb *WebhookContextBase) SaveAppUser(appUserID int64, appUserEntity AppUser) error {
	return whcb.AppUserStore.SaveAppUser(appUserID, appUserEntity)
}


func (whcb *WebhookContextBase) SetChatEntity(chatEntity BotChat) {
	whcb.chatEntity = chatEntity
}

func (whcb *WebhookContextBase) ChatEntity(whc WebhookContext) (BotChat, error) {
	if whcb.chatEntity == nil {
		err := whcb.getChatEntityBase(whc)
		if err != nil {
			return nil, err
		}
	}
	return whcb.chatEntity, nil
}

func (whcb *WebhookContextBase) GetOrCreateBotUserEntityBase() (BotUser, error) {
	botUser, err := whcb.GetBotUserById(whcb.GetSender().GetID())
	if err != nil {
		return nil, err
	}
	logger := whcb.GetLogger()
	if botUser == nil {
		logger.Infof("Bot user entity not found, creating a new one...")
		botUser, err = whcb.CreateBotUser(whcb.GetSender())
		if err != nil {
			return nil, err
		}
		logger.Infof("Bot user entity created")
	} else {
		logger.Infof("Found existing bot user entity")
	}
	return botUser, err
}

func (whcb *WebhookContextBase) getChatEntityBase(whc WebhookContext) error {
	logger := whcb.GetLogger()
	if whcb.HasChatEntity() {
		logger.Warningf("Duplicate call of func (whc *bot.WebhookContext) _getChat()")
		return nil
	}

	botChatID := whc.BotChatID()
	logger.Infof("botChatID: %v", botChatID)
	botChatEntity, err := whcb.BotChatStore.GetBotChatEntityById(botChatID)
	switch err {
	case nil: // Nothing to do
		logger.Debugf("Loaded botChatEntity: %v", botChatEntity)
	case ErrEntityNotFound: //TODO: Should be this moved to DAL?
		err = nil
		logger.Infof("BotChat not found, first check for bot user entity...")
		botUser, err := whcb.GetOrCreateBotUserEntityBase()
		if err != nil {
			return err
		}
		botChatEntity = whcb.BotChatStore.NewBotChatEntity(botChatID, botUser.GetAppUserID(), botChatID, botUser.IsAccessGranted())
	default:
		return err
	}
	logger.Debugf(`chatEntity.PreferredLanguage: %v, whc.locale.Code5: %v, chatEntity.PreferredLanguage != """ && whc.locale.Code5 != chatEntity.PreferredLanguage: %v`, botChatEntity.GetPreferredLanguage(), whc.Locale().Code5, botChatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != botChatEntity.GetPreferredLanguage())
	if botChatEntity.GetPreferredLanguage() != "" && whc.Locale().Code5 != botChatEntity.GetPreferredLanguage() {
		err = whc.SetLocale(botChatEntity.GetPreferredLanguage())
		if err == nil {
			logger.Debugf("whc.locale cahged to: %v", whc.Locale)
		} else {
			logger.Errorf("Failed to set locate: %v")
		}
	}
	whcb.chatEntity = botChatEntity
	return err
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
	if whcb.locale.Code5 == "" {
		return whcb.BotContext.BotSettings.Locale
	}
	return whcb.locale
}

func (whcb *WebhookContextBase) SetLocale(code5 string) error {
	locale, err := whcb.AppContext.SupportedLocales().GetLocaleByCode5(code5)
	if err == nil {
		whcb.locale = locale
	}
	return err
}
