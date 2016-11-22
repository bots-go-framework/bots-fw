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
)

type WebhookContextBase struct {
	//w          http.ResponseWriter
	r             *http.Request
	c             context.Context
	logger        strongo.Logger
	botAppContext BotAppContext
	BotContext    BotContext
	botPlatform   BotPlatform
	input         WebhookInput

	locale        strongo.Locale

	//update      tgbotapi.Update
	chatEntity    BotChat

	BotUserKey    *datastore.Key
	appUser       BotAppUser
	strongo.Translator
	//Locales    strongo.LocalesProvider

	BotCoreStores

	gaMeasurement *measurement.BufferedSender
}

func (whc *WebhookContextBase) BotChatID() (chatID string) {
	input := whc.Input()
	if chat := input.Chat(); chat != nil {
		return chat.GetID()
	}
	switch input.(type) {
	case WebhookCallbackQuery:
		callbackQuery := input.(WebhookCallbackQuery)
		data := callbackQuery.GetData()
		if strings.Contains(data, "chat=") {
			c := whc.Context()
			values, err := url.ParseQuery(data)
			if err != nil {
				whc.Logger().Errorf(c, "Failed to GetData() from webhookInput.InputCallbackQuery()")
				return ""
			}
			chatID = values.Get("chat")
		}
	default:
		whc.logger.Warningf(whc.c, "*.WebhookContextBaseBotChatID(): Unhandled input type: %T", input)
	}

	return chatID
}

func (whc *WebhookContextBase) AppUserIntID() (appUserIntID int64) {
	if chatEntity := whc.ChatEntity(); chatEntity != nil {
		appUserIntID = chatEntity.GetAppUserIntID()
	}
	if appUserIntID == 0 {
		botUser, err := whc.GetOrCreateBotUserEntityBase()
		if err != nil {
			panic(fmt.Sprintf("Failed to get bot user entity: %v", err))
		}
		appUserIntID = botUser.GetAppUserIntID()
	}
	return
}


func (whc *WebhookContextBase) GetAppUser() (BotAppUser, error) {
	appUserID := whc.AppUserIntID()
	appUser := whc.BotAppContext().NewBotAppUserEntity()
	err := whc.BotAppUserStore.GetAppUserByID(whc.Context(), appUserID, appUser)
	return appUser, err
}


func (whcb *WebhookContextBase) ExecutionContext() strongo.ExecutionContext {
	return whcb
}

func (whcb *WebhookContextBase) BotAppContext() BotAppContext {
	return whcb.botAppContext
}

func NewWebhookContextBase(r *http.Request, botAppContext BotAppContext, botPlatform BotPlatform, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores, gaMeasurement *measurement.BufferedSender) *WebhookContextBase {
	whcb := WebhookContextBase{
		r:             r,
		c:             appengine.NewContext(r),
		gaMeasurement: gaMeasurement,
		logger:        botContext.BotHost.Logger(r),
		botAppContext: botAppContext,
		botPlatform:   botPlatform,
		BotContext:    botContext,
		input:         webhookInput,
		BotCoreStores: botCoreStores,
	}
	whcb.Translator = botAppContext.GetTranslator(whcb.c, whcb.logger)
	return &whcb
}

func (whcb *WebhookContextBase)  Input() WebhookInput {
	return whcb.input
}

func (whcb *WebhookContextBase)  Chat() WebhookChat {
	return whcb.input.Chat()
}

func (whcb *WebhookContextBase)  GetRecipient() WebhookRecipient {
	return whcb.input.GetRecipient()
}

func (whcb *WebhookContextBase)  GetSender() WebhookSender {
	return whcb.input.GetSender()
}

func (whcb *WebhookContextBase)  GetTime() time.Time {
	return whcb.input.GetTime()
}

func (whcb *WebhookContextBase)  InputType() WebhookInputType {
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

func (whcb *WebhookContextBase) Logger() strongo.Logger {
	return whcb.logger
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
	return whcb.BotContext.BotHost.GetHttpClient(whcb.r)
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
	if whcb.BotChatID() == "" {
		whcb.logger.Debugf(whcb.c, "whcb.BotChatID() is empty string")
		return nil
	}
	if whcb.chatEntity == nil {
		if err := whcb.loadChatEntityBase(); err != nil {
			panic(errors.Wrap(err, "Failed to call whcb.getChatEntityBase()"))
		}
	}
	return whcb.chatEntity
}

func (whcb *WebhookContextBase) GetOrCreateBotUserEntityBase() (BotUser, error) {
	logger := whcb.Logger()
	c := whcb.Context()
	logger.Debugf(c, "GetOrCreateBotUserEntityBase()")
	sender := whcb.input.GetSender()
	botUserID := sender.GetID()
	botUser, err := whcb.GetBotUserById(c, botUserID)
	if err != nil {
		return nil, err
	}
	if botUser == nil {
		logger.Infof(c, "Bot user entity not found, creating a new one...")
		botUser, err = whcb.CreateBotUser(c, whcb.GetBotCode(), sender)
		if err != nil {
			logger.Errorf(c, "Failed to create bot user: %v", err)
			return nil, err
		}
		logger.Infof(c, "Bot user entity created")

		whcb.gaMeasurement.Queue(whcb.GaEvent("users", "user-created")) //TODO: Should be outside

		whcb.gaMeasurement.Queue(whcb.GaEventWithLabel("users", "messenger-linked", whcb.botPlatform.Id())) // TODO: Should be outside

		if whcb.GetBotSettings().Mode == Production {
			gaEvent := measurement.NewEvent("bot-users", "bot-user-created", whcb.GaCommon())
			gaEvent.Label = fmt.Sprintf("%v", botUserID)
			whcb.GaMeasurement().Queue(gaEvent)
		}
	} else {
		logger.Infof(c, "Found existing bot user entity")
	}
	return botUser, err
}

func (whcb *WebhookContextBase) loadChatEntityBase() error {
	logger := whcb.Logger()
	c := whcb.Context()
	if whcb.HasChatEntity() {
		logger.Warningf(c, "Duplicate call of func (whc *bot.WebhookContext) _getChat()")
		return nil
	}

	botChatID := whcb.BotChatID()
	logger.Infof(c, "loadChatEntityBase(): botChatID: %v", botChatID)
	botID := whcb.GetBotCode()
	botChatStore := whcb.BotChatStore
	if botChatStore == nil {
		panic("botChatStore == nil")
	}
	botChatEntity, err := botChatStore.GetBotChatEntityByID(c, botID, botChatID)
	switch err {
	case nil: // Nothing to do
		logger.Debugf(c, "GetBotChatEntityByID() returned nil")
	case ErrEntityNotFound: //TODO: Should be this moved to DAL?
		err = nil
		logger.Infof(c, "BotChat not found, first check for bot user entity...")
		botUser, err := whcb.GetOrCreateBotUserEntityBase()
		if err != nil {
			return err
		}

		botChatEntity = whcb.BotChatStore.NewBotChatEntity(c, whcb.GetBotCode(), botChatID, botUser.GetAppUserIntID(), botChatID, botUser.IsAccessGranted())

		if whcb.GetBotSettings().Mode == Production {
			gaEvent := measurement.NewEvent("bot-chats", "bot-chat-created", whcb.GaCommon())
			gaEvent.Label = fmt.Sprintf("%v", botChatID)
			whcb.GaMeasurement().Queue(gaEvent)
		}

	default:
		return err
	}

	logger.Debugf(c, `chatEntity.PreferredLanguage: %v, whc.locale.Code5: %v, chatEntity.PreferredLanguage != """ && whc.locale.Code5 != chatEntity.PreferredLanguage: %v`,
		botChatEntity.GetPreferredLanguage(), whcb.Locale().Code5, botChatEntity.GetPreferredLanguage() != "" && whcb.Locale().Code5 != botChatEntity.GetPreferredLanguage())

	if botChatEntity.GetPreferredLanguage() != "" && whcb.Locale().Code5 != botChatEntity.GetPreferredLanguage() {
		err = whcb.SetLocale(botChatEntity.GetPreferredLanguage())
		if err == nil {
			logger.Debugf(c, "whc.locale changed to: %v", whcb.Locale().Code5)
		} else {
			logger.Errorf(c, "Failed to set locate: %v")
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

func (whcb *WebhookContextBase) NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot {
	return MessageFromBot{Text: fmt.Sprintf(whcb.Translate(messageCode), a...), Format: MessageFormatHTML}
}

func (whcb *WebhookContextBase) MessageText() string {
	if tm, ok := whcb.Input().(WebhookTextMessage); ok {
		return tm.Text()
	}
	return ""
}

func (whcb *WebhookContextBase) NewMessage(text string) MessageFromBot {
	return MessageFromBot{Text: text, Format: MessageFormatHTML}
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
		whcb.logger.Errorf(whcb.c, "WebhookContextBase.SetLocate() - %v", err)
		return err
	}
	whcb.locale = locale
	return nil
}
