package bots

import (
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
	"strconv"
	"strings"
	"time"
	"net/url"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"github.com/strongo/app/db"
)

type WebhookContextBase struct {
	//w          http.ResponseWriter
	r             *http.Request
	c             context.Context
	botAppContext BotAppContext
	BotContext    BotContext
	botPlatform   BotPlatform
	input         WebhookInput

	locale strongo.Locale

	//update      tgbotapi.Update
	chatID  string
	chatEntity BotChat

	BotUserKey *datastore.Key
	appUser    BotAppUser
	strongo.Translator
	//Locales    strongo.LocalesProvider

	BotCoreStores

	gaMeasurement *measurement.BufferedSender
}

func (whcb *WebhookContextBase) SetChatID(v string) {
	whcb.chatID = v
}

func (whcb *WebhookContextBase) LogRequest() {
	whcb.input.LogRequest()
}

func (whcb *WebhookContextBase) RunInTransaction(c context.Context, f func(c context.Context) error, options db.RunOptions) error {
	return whcb.BotContext.BotHost.DB().RunInTransaction(c, f, options)
}

func (whcb *WebhookContextBase) Request() *http.Request {
	return whcb.r
}

func (whcb *WebhookContextBase) Environment() strongo.Environment {
	return whcb.BotContext.BotSettings.Env
}

func (whcb *WebhookContextBase) MustBotChatID() (chatID string) {
	var err error
	if chatID, err = whcb.BotChatID(); err != nil {
		panic(err)
	} else if chatID == "" {
		panic("BotChatID() returned an empty string")
	}
	return
}

func (whcb *WebhookContextBase) BotChatID() (string, error) {
	if whcb.chatID != "" {
		return whcb.chatID, nil
	}
	input := whcb.Input()
	if chat := input.Chat(); chat != nil {
		whcb.chatID = chat.GetID()
		return whcb.chatID, nil
	}
	switch input.(type) {
	case WebhookCallbackQuery:
		callbackQuery := input.(WebhookCallbackQuery)
		data := callbackQuery.GetData()
		if strings.Contains(data, "chat=") {
			if values, err := url.ParseQuery(data); err != nil {
				return "", errors.WithMessage(err, "Failed to GetData() from webhookInput.InputCallbackQuery()")
			} else {
				whcb.chatID = values.Get("chat")
			}
		}
	case WebhookInlineQuery:
		// pass
	default:
		whcb.LogRequest()
		log.Debugf(whcb.c, "BotChatID(): *.WebhookContextBaseBotChatID(): Unhandled input type: %T", input)
	}

	return whcb.chatID, nil
}

func (whcb *WebhookContextBase) AppUserIntID() (appUserIntID int64) {
	if chatEntity := whcb.ChatEntity(); chatEntity != nil {
		appUserIntID = chatEntity.GetAppUserIntID()
	}
	if appUserIntID == 0 {
		botUser, err := whcb.GetOrCreateBotUserEntityBase()
		if err != nil {
			panic(fmt.Sprintf("Failed to get bot user entity: %v", err))
		}
		appUserIntID = botUser.GetAppUserIntID()
	}
	return
}

func (whcb *WebhookContextBase) GetAppUser() (BotAppUser, error) { // TODO: Can/should this be cached?
	appUserID := whcb.AppUserIntID()
	appUser := whcb.BotAppContext().NewBotAppUserEntity()
	err := whcb.BotAppUserStore.GetAppUserByID(whcb.Context(), appUserID, appUser)
	return appUser, err
}

func (whcb *WebhookContextBase) ExecutionContext() strongo.ExecutionContext {
	return whcb
}

func (whcb *WebhookContextBase) BotAppContext() BotAppContext {
	return whcb.botAppContext
}

func NewWebhookContextBase(
	r *http.Request,
	botAppContext BotAppContext,
	botPlatform BotPlatform,
	botContext BotContext,
	webhookInput WebhookInput,
	botCoreStores BotCoreStores,
	gaMeasurement *measurement.BufferedSender,
) *WebhookContextBase {
	whcb := WebhookContextBase{
		r:             r,
		c:             appengine.NewContext(r),
		gaMeasurement: gaMeasurement,
		botAppContext: botAppContext,
		botPlatform:   botPlatform,
		BotContext:    botContext,
		input:         webhookInput,
		BotCoreStores: botCoreStores,
	}
	whcb.Translator = botAppContext.GetTranslator(whcb.c)
	return &whcb
}

func (whcb *WebhookContextBase) Input() WebhookInput {
	return whcb.input
}

func (whcb *WebhookContextBase) Chat() WebhookChat {
	return whcb.input.Chat()
}

func (whcb *WebhookContextBase) GetRecipient() WebhookRecipient {
	return whcb.input.GetRecipient()
}

func (whcb *WebhookContextBase) GetSender() WebhookSender {
	return whcb.input.GetSender()
}

func (whcb *WebhookContextBase) GetTime() time.Time {
	return whcb.input.GetTime()
}

func (whcb *WebhookContextBase) InputType() WebhookInputType {
	return whcb.input.InputType()
}

func (whcb *WebhookContextBase) GaMeasurement() *measurement.BufferedSender {
	return whcb.gaMeasurement
}

func (whcb *WebhookContextBase) GaCommon() measurement.Common {
	if whcb.chatEntity != nil {
		c := whcb.Context()
		return measurement.Common{
			UserID:        strconv.FormatInt(whcb.chatEntity.GetAppUserIntID(), 10),
			UserLanguage:  strings.ToLower(whcb.chatEntity.GetPreferredLanguage()),
			ClientID:      whcb.chatEntity.GetGaClientID().String(),
			ApplicationID: fmt.Sprintf("bot.%v.%v", whcb.botPlatform.Id(), whcb.GetBotCode()),
			UserAgent:     fmt.Sprintf("%v bot (%v:%v) %v", whcb.botPlatform.Id(), appengine.AppID(c), appengine.VersionID(c), whcb.r.Host),
			DataSource:    "bot",
		}
	}
	return measurement.Common{
		DataSource: "bot",
		ClientID:   "c7ea15eb-3333-4d47-a002-9d1a14996371",
	}
}

func (whcb *WebhookContextBase) GaEvent(category, action string) measurement.Event {
	return measurement.NewEvent(category, action, whcb.GaCommon())
}

func (whcb *WebhookContextBase) GaEventWithLabel(category, action, label string) measurement.Event {
	return measurement.NewEventWithLabel(category, action, label, whcb.GaCommon())
}

func (whcb *WebhookContextBase) BotPlatform() BotPlatform {
	return whcb.botPlatform
}

func (whcb *WebhookContextBase) GetBotSettings() BotSettings {
	return whcb.BotContext.BotSettings
}

func (whcb *WebhookContextBase) GetBotCode() string {
	return whcb.BotContext.BotSettings.Code
}

func (whcb *WebhookContextBase) GetBotToken() string {
	return whcb.BotContext.BotSettings.Token
}

func (whcb *WebhookContextBase) Translate(key string, args ...interface{}) string {
	return whcb.Translator.Translate(key, whcb.Locale().Code5, args...)
}

func (whcb *WebhookContextBase) TranslateNoWarning(key string, args ...interface{}) string {
	return whcb.Translator.TranslateNoWarning(key, whcb.locale.Code5, args...)
}

func (whcb *WebhookContextBase) GetHttpClient() *http.Client {
	return whcb.BotContext.BotHost.GetHttpClient(whcb.c)
}

func (whcb *WebhookContextBase) HasChatEntity() bool {
	return whcb.chatEntity != nil
}

func (whcb *WebhookContextBase) SaveAppUser(appUserID int64, appUserEntity BotAppUser) error {
	return whcb.BotAppUserStore.SaveAppUser(whcb.Context(), appUserID, appUserEntity)
}

func (whcb *WebhookContextBase) SetChatEntity(chatEntity BotChat) {
	whcb.chatEntity = chatEntity
}

func (whcb *WebhookContextBase) ChatEntity() BotChat {
	if whcb.chatEntity != nil {
		return whcb.chatEntity
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		panic(errors.WithMessage(err, "failed to call whcb.BotChatID()"))
	}
	if chatID == "" {
		log.Debugf(whcb.c, "whcb.BotChatID() is empty string")
		return nil
	}
	if err := whcb.loadChatEntityBase(); err != nil {
		panic(errors.Wrap(err, "Failed to call whcb.getChatEntityBase()"))
	}
	return whcb.chatEntity
}

func (whcb *WebhookContextBase) GetOrCreateBotUserEntityBase() (BotUser, error) {
	c := whcb.Context()
	log.Debugf(c, "GetOrCreateBotUserEntityBase()")
	sender := whcb.input.GetSender()
	botUserID := sender.GetID()
	botUser, err := whcb.GetBotUserById(c, botUserID)
	if err != nil {
		return nil, err
	}
	if botUser == nil {
		log.Infof(c, "Bot user entity not found, creating a new one...")
		botUser, err = whcb.CreateBotUser(c, whcb.GetBotCode(), sender)
		if err != nil {
			log.Errorf(c, "Failed to create bot user: %v", err)
			return nil, err
		}
		log.Infof(c, "Bot user entity created")

		whcb.gaMeasurement.Queue(whcb.GaEvent("users", "user-created")) //TODO: Should be outside

		whcb.gaMeasurement.Queue(whcb.GaEventWithLabel("users", "messenger-linked", whcb.botPlatform.Id())) // TODO: Should be outside

		if whcb.GetBotSettings().Env == strongo.EnvProduction {
			gaEvent := measurement.NewEvent("bot-users", "bot-user-created", whcb.GaCommon())
			gaEvent.Label = whcb.botPlatform.Id()
			whcb.GaMeasurement().Queue(gaEvent)
		}
	} else {
		log.Infof(c, "Found existing bot user entity")
	}
	return botUser, err
}

func (whcb *WebhookContextBase) loadChatEntityBase() error {
	c := whcb.Context()
	if whcb.HasChatEntity() {
		log.Warningf(c, "Duplicate call of func (whc *bot.WebhookContext) _getChat()")
		return nil
	}

	botChatID, err := whcb.BotChatID()
	if err != nil {
		return errors.WithMessage(err, "Failed to call whcb.BotChatID()")
	}

	log.Debugf(c, "loadChatEntityBase(): botChatID: %v", botChatID)
	botID := whcb.GetBotCode()
	botChatStore := whcb.BotChatStore
	if botChatStore == nil {
		panic("botChatStore == nil")
	}
	botChatEntity, err := botChatStore.GetBotChatEntityByID(c, botID, botChatID)
	switch err {
	case nil: // Nothing to do
		log.Debugf(c, "GetBotChatEntityByID() returned => AwaitingReplyTo: %v", botChatEntity.GetAwaitingReplyTo())
	case ErrEntityNotFound: //TODO: Should be this moved to DAL?
		err = nil
		log.Infof(c, "BotChat not found, first check for bot user entity...")
		botUser, err := whcb.GetOrCreateBotUserEntityBase()
		if err != nil {
			return err
		}
		botChatEntity = whcb.BotChatStore.NewBotChatEntity(c, whcb.GetBotCode(), whcb.input.Chat(), botUser.GetAppUserIntID(), botChatID, botUser.IsAccessGranted())

		if whcb.GetBotSettings().Env == strongo.EnvProduction {
			gaEvent := measurement.NewEvent("bot-chats", "bot-chat-created", whcb.GaCommon())
			gaEvent.Label = whcb.botPlatform.Id()
			whcb.GaMeasurement().Queue(gaEvent)
		}
	default:
		return err
	}

	if !botChatEntity.IsGroupChat() {
		if sender := whcb.input.GetSender(); sender != nil {
			if languageCode := sender.GetLanguage(); languageCode != "" {
				botChatEntity.AddClientLanguage(languageCode)
			}
		}
	}

	log.Debugf(c, `chatEntity.PreferredLanguage: %v, whc.locale.Code5: %v, chatEntity.PreferredLanguage != """ && whc.locale.Code5 != chatEntity.PreferredLanguage: %v`,
		botChatEntity.GetPreferredLanguage(), whcb.Locale().Code5, botChatEntity.GetPreferredLanguage() != "" && whcb.Locale().Code5 != botChatEntity.GetPreferredLanguage())

	if botChatEntity.GetPreferredLanguage() != "" && whcb.Locale().Code5 != botChatEntity.GetPreferredLanguage() {
		err = whcb.SetLocale(botChatEntity.GetPreferredLanguage())
		if err == nil {
			log.Debugf(c, "whc.locale changed to: %v", whcb.Locale().Code5)
		} else {
			log.Errorf(c, "Failed to set locate: %v")
		}
	}
	whcb.chatEntity = botChatEntity
	return err
}

func (whcb *WebhookContextBase) AppUserEntity() BotAppUser {
	return whcb.appUser
}

func (whcb *WebhookContextBase) Context() context.Context {
	return whcb.c
}

func (whcb *WebhookContextBase) NewMessageByCode(messageCode string, a ...interface{}) (m MessageFromBot) {
	return whcb.NewMessage(fmt.Sprintf(whcb.Translate(messageCode), a...))
}

func (whcb *WebhookContextBase) MessageText() string {
	if tm, ok := whcb.Input().(WebhookTextMessage); ok {
		return tm.Text()
	}
	return ""
}

func (whcb *WebhookContextBase) NewMessage(text string) (m MessageFromBot) {
	m.Text = text
	m.Format = MessageFormatHTML
	return
}

func (whcb WebhookContextBase) Locale() strongo.Locale {
	if whcb.locale.Code5 == "" {
		return whcb.BotContext.BotSettings.Locale
	}
	return whcb.locale
}

func (whcb *WebhookContextBase) SetLocale(code5 string) error {
	locale, err := whcb.botAppContext.SupportedLocales().GetLocaleByCode5(code5)
	if err != nil {
		log.Errorf(whcb.c, "WebhookContextBase.SetLocate() - %v", err)
		return err
	}
	whcb.locale = locale
	return nil
}
