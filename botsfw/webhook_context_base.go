package botsfw

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfw/botsdal"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/gamp"
	"github.com/strongo/i18n"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var _ WebhookContext = (*whContextDummy)(nil)

// whContextDummy is a dummy implementation of WebhookContext interface
// It exists only to check what is NOT implemented by WebhookContextBase
type whContextDummy struct {
	*WebhookContextBase
}

func (w whContextDummy) NewEditMessage(text string, format MessageFormat) (MessageFromBot, error) {
	panic(fmt.Sprintf("must be implemented in platform specific code: text=%s, format=%v", text, format))
}

func (w whContextDummy) UpdateLastProcessed(chatEntity botsfwmodels.BotChatData) error {
	panic(fmt.Sprintf("implement me in WebhookContextBase - UpdateLastProcessed(chatEntity=%v)", chatEntity))
}

func (w whContextDummy) AppUserData() (botsfwmodels.AppUserData, error) {
	panic("implement me in WebhookContextBase") //TODO
}

func (w whContextDummy) IsNewerThen(chatEntity botsfwmodels.BotChatData) bool {
	panic(fmt.Sprintf("implement me in WebhookContextBase - IsNewerThen(chatEntity=%v)", chatEntity))
}

func (w whContextDummy) Responder() WebhookResponder {
	//TODO implement me
	panic("implement me")
}

// WebhookContextBase provides base implementation of WebhookContext interface
// TODO: Document purpose of a dedicated base struct (e.g. example of usage by developers)
type WebhookContextBase struct {
	//w http.ResponseWriter
	r             *http.Request
	c             context.Context
	botAppContext BotAppContext
	botContext    BotContext // TODO: rename to something strongo
	botPlatform   BotPlatform
	input         WebhookInput
	//recordsMaker        botsfwmodels.BotRecordsMaker
	recordsFieldsSetter BotRecordsFieldsSetter

	isInGroup func() bool

	getLocaleAndChatID func() (locale, chatID string, err error) // TODO: Document why we need to pass context. Is it to support transactions?

	locale i18n.Locale

	// At the moment, there is no reason to expose botChat record publicly
	// If there is some it should be documented with a use case
	botChat record.DataWithID[string, botsfwmodels.BotChatData]
	botUser record.DataWithID[string, botsfwmodels.BotUserData] // Telegram user ID is an integer, but we keep it as a string for consistency & simplicity.

	isLoadingChatData bool // TODO: This smells bad. Needs refactoring?
	isLoadingUserData bool // TODO: This smells bad. Needs refactoring?

	//
	appUserData botsfwmodels.AppUserData

	translator
	//Locales    strongoapp.LocalesProvider

	//dal botsfwdal.DataAccess
	db dal.DB
	tx dal.ReadwriteTransaction

	gaContext gaContext
}

func (whcb *WebhookContextBase) DB() dal.DB {
	return whcb.db
}

// Tx returns a transaction that is used to read/write botChat & bot user data
func (whcb *WebhookContextBase) Tx() dal.ReadwriteTransaction {
	return whcb.tx
}

func (whcb *WebhookContextBase) RecordsFieldsSetter() BotRecordsFieldsSetter {
	return whcb.recordsFieldsSetter
}

//func (whcb *WebhookContextBase) Store() botsfwdal.DataAccess {
//	return whcb.dal
//}

func (whcb *WebhookContextBase) BotContext() BotContext {
	return whcb.botContext
}

// SetChatID sets botChat ID - TODO: Should it be private?
func (whcb *WebhookContextBase) SetChatID(chatID string) {
	whcb.botChat.ID = chatID
	//whcb.botChat.Key = botsdal.newChatKey(whcb.botPlatform.ID(), whcb.botContext.BotSettings.Code, chatID)
}

// LogRequest logs request data to logging system
func (whcb *WebhookContextBase) LogRequest() {
	whcb.input.LogRequest()
}

// // RunReadwriteTransaction starts a transaction. This needed to coordinate application & framework changes.
//func (whcb *WebhookContextBase) RunReadwriteTransaction(c context.Context, f func(ctx context.Context)) error {
//	db, err := whcb.botContext.BotHost.DB(c)
//	if err != nil {
//		return err
//	}
//	return db.RunReadwriteTransaction(c, f, options...)
//}

//
//// RunReadonlyTransaction starts a readonly transaction.
//func (whcb *WebhookContextBase) RunReadonlyTransaction(c context.Context, f dal.ROTxWorker, options ...dal.TransactionOption) error {
//	db, err := whcb.botContext.BotHost.DB(c)
//	if err != nil {
//		return err
//	}
//	return db.RunReadonlyTransaction(c, f, options...)
//}

// IsInTransaction detects if request is within a transaction
func (whcb *WebhookContextBase) IsInTransaction(context.Context) bool {
	panic("not implemented")
	//return whcb.botContext.BotHost.DB().IsInTransaction(c)
}

// NonTransactionalContext creates a non transaction context for operations that needs to be executed outside of transaction.
func (whcb *WebhookContextBase) NonTransactionalContext(context.Context) context.Context {
	panic("not implemented")
	//return whcb.botContext.BotHost.DB().NonTransactionalContext(tc)
}

// Request returns reference to current HTTP request
func (whcb *WebhookContextBase) Request() *http.Request {
	return whcb.r
}

// Environment defines current environment (PROD, DEV, LOCAL, etc)
func (whcb *WebhookContextBase) Environment() string {
	return whcb.botContext.BotSettings.Env
}

// MustBotChatID returns bot botChat ID and panic if missing it
func (whcb *WebhookContextBase) MustBotChatID() (chatID string) {
	var err error
	if chatID, err = whcb.BotChatID(); err != nil {
		panic(err)
	} else if chatID == "" {
		panic("BotChatID() returned an empty string")
	}
	return
}

// BotChatID returns bot botChat ID
func (whcb *WebhookContextBase) BotChatID() (botChatID string, err error) {
	if whcb.botChat.ID != "" {
		return whcb.botChat.ID, nil
	}
	//log.Debugf(whcb.c, "*WebhookContextBase.BotChatID()")

	input := whcb.Input()
	if botChatID, err = input.BotChatID(); err != nil {
		return
	} else if botChatID != "" {
		whcb.SetChatID(botChatID)
		return whcb.botChat.ID, nil
	}
	if whcb.getLocaleAndChatID != nil {
		if _, botChatID, err = whcb.getLocaleAndChatID(); err != nil {
			return
		}
		if botChatID != "" {
			whcb.SetChatID(botChatID)
			return whcb.botChat.ID, nil
		}
	}
	switch input := input.(type) {
	case WebhookCallbackQuery:
		data := input.GetData()
		if strings.Contains(data, "botChat=") {
			values, err := url.ParseQuery(data)
			if err != nil {
				return "", fmt.Errorf("failed to GetData() from webhookInput.InputCallbackQuery(): %w", err)
			}
			chatID := values.Get("botChat")
			whcb.SetChatID(chatID)
		}
	case WebhookInlineQuery:
		// pass
	case WebhookChosenInlineResult:
		// pass
	default:
		whcb.LogRequest()
		log.Debugf(whcb.c, "BotChatID(): *.WebhookContextBaseBotChatID(): Unhandled input type: %T", input)
	}

	return whcb.botChat.ID, nil
}

// AppUserInt64ID Deprecate: use AppUserID() instead
//func (whcb *WebhookContextBase) AppUserInt64ID() (appUserID int64) {
//	if s := whcb.AppUserID(); s != "" {
//		var err error
//		if appUserID, err = strconv.ParseInt(s, 10, 64); err != nil {
//			panic(fmt.Errorf("failed to parse app user ID %v: %w", s, err))
//		}
//	}
//	return appUserID
//}

// AppUserID return current app user ID as a string. AppUserIntID() is deprecated.
func (whcb *WebhookContextBase) AppUserID() (appUserID string) {
	if !whcb.isLoadingChatData && !whcb.isLoadingUserData {
		whcb.isInGroup()
		if chatData := whcb.ChatData(); chatData != nil {
			appUserID = chatData.GetAppUserID()
		}
	}
	if appUserID == "" && !whcb.isLoadingUserData {
		if botUser, err := whcb.getOrCreateBotUserData(); err != nil {
			if !dal.IsNotFound(err) {
				panic(fmt.Errorf("failed to get bot user entity: %w", err))
			}
		} else {
			appUserID = botUser.GetAppUserID()
		}
	}
	return
}

func (whcb *WebhookContextBase) BotUser() (botUser record.DataWithID[string, botsfwmodels.BotUserData], err error) {
	if whcb.botUser.Data != nil {
		return whcb.botUser, nil
	}
	botID := whcb.botContext.BotSettings.ID
	platformID := whcb.botContext.BotSettings.Profile.ID()
	botUserID := whcb.GetBotUserID()
	whcb.botUser, err = botsdal.GetBotUser(whcb.c, whcb.tx, platformID, botID, botUserID, whcb.botContext.BotSettings.Profile.NewBotUserData)
	return whcb.botUser, err
}

// GetAppUser loads information about current app user from persistent storage
func (whcb *WebhookContextBase) GetAppUser() (botsfwmodels.AppUserData, error) { // TODO: Can/should this be cached?
	appUserID := whcb.AppUserID()
	appUser, err := whcb.BotContext().BotSettings.GetAppUserByID(whcb.c, whcb.tx, appUserID)
	if err != nil {
		return nil, err
	}
	return appUser.Data, err
}

// ExecutionContext returns an execution context for strongo app
func (whcb *WebhookContextBase) ExecutionContext() ExecutionContext {
	return whcb
}

// BotAppContext returns bot app context
func (whcb *WebhookContextBase) BotAppContext() BotAppContext {
	return whcb.botAppContext
}

// IsInGroup signals if the bot request is send within group botChat
func (whcb *WebhookContextBase) IsInGroup() bool {
	return whcb.isInGroup()
}

// NewWebhookContextBase creates base bot context
func NewWebhookContextBase(
	args CreateWebhookContextArgs,
	botPlatform BotPlatform,
	recordsFieldsSetter BotRecordsFieldsSetter, // TODO: Should it be a member of BotContext?
	isInGroup func() bool,
	getLocaleAndChatID func(c context.Context) (locale, chatID string, err error),
) (*WebhookContextBase, error) {
	if args.HttpRequest == nil {
		panic("args.HttpRequest == nil")
	}
	c := args.BotContext.BotHost.Context(args.HttpRequest)
	whcb := WebhookContextBase{
		r:  args.HttpRequest,
		c:  c,
		tx: args.Tx,
		getLocaleAndChatID: func() (locale, chatID string, err error) {
			return getLocaleAndChatID(c)
		},
		botAppContext: args.AppContext,
		botPlatform:   botPlatform,
		botContext:    args.BotContext,
		input:         args.WebhookInput,
		isInGroup:     isInGroup,
		//dal:                 botCoreStores,
		recordsFieldsSetter: recordsFieldsSetter,
	}
	whcb.gaContext = gaContext{
		whcb:          &whcb,
		gaMeasurement: args.GaMeasurement,
	}
	if isInGroup() && whcb.getLocaleAndChatID != nil {
		if locale, chatID, err := whcb.getLocaleAndChatID(); err != nil {
			panic(err)
		} else {
			if chatID != "" {
				whcb.SetChatID(chatID)
			}
			if locale != "" {
				if err := whcb.SetLocale(locale); err != nil {
					log.Errorf(c, "Failed to set Locale: %v", err)
				}
			}
		}
	}
	whcb.translator = translator{
		localeCode5: func() string {
			return whcb.locale.Code5
		},
		Translator: args.AppContext.GetTranslator(whcb.c),
	}
	return &whcb, nil
}

// Input returns webhook input
func (whcb *WebhookContextBase) Input() WebhookInput {
	return whcb.input
}

// Chat returns webhook botChat
func (whcb *WebhookContextBase) Chat() WebhookChat { // TODO: remove
	return whcb.input.Chat()
}

// GetRecipient returns receiver of the message
func (whcb *WebhookContextBase) GetRecipient() WebhookRecipient { // TODO: remove
	return whcb.input.GetRecipient()
}

// GetSender returns sender of the message
func (whcb *WebhookContextBase) GetSender() WebhookSender { // TODO: remove
	return whcb.input.GetSender()
}

// GetTime returns time of the message
func (whcb *WebhookContextBase) GetTime() time.Time { // TODO: remove
	return whcb.input.GetTime()
}

// InputType returns input type
func (whcb *WebhookContextBase) InputType() WebhookInputType { // TODO: remove
	return whcb.input.InputType()
}

// GaMeasurement returns a provider to send information to Google Analytics
func (gac gaContext) GaMeasurement() GaQueuer {
	return gac.gaMeasurement
}

type gaContext struct {
	whcb          *WebhookContextBase
	gaMeasurement GaQueuer
}

// GA provides interface to Google Analytics
func (whcb *WebhookContextBase) GA() GaContext {
	return whcb.gaContext
}

func (gac gaContext) Queue(message gamp.Message) error {
	if gac.gaMeasurement == nil { // TODO: not good :(
		return nil
	}
	if message.GetTrackingID() == "" {
		message.SetTrackingID(gac.whcb.GetBotSettings().GAToken)
		if message.GetTrackingID() == "" {
			return fmt.Errorf("gaContext.Queue(%v): %w", message, gamp.ErrNoTrackingID)
		}
	}
	return gac.gaMeasurement.Queue(message)
}

//	func (gac gaContext) Flush() error {
//		return gac.gaMeasurement.
//	}
//
// GaCommon creates context for Google Analytics
func (gac gaContext) GaCommon() gamp.Common {
	whcb := gac.whcb
	if whcb.botChat.Record.Exists() {
		return gamp.Common{
			UserID:       whcb.botChat.Data.GetAppUserID(),
			UserLanguage: strings.ToLower(whcb.botChat.Data.GetPreferredLanguage()),
			//ClientID:      whcb.chatData.GetGaClientID(), // TODO: Restore feature
			ApplicationID: fmt.Sprintf("bot.%v.%v", whcb.botPlatform.ID(), whcb.GetBotCode()),
			UserAgent:     fmt.Sprintf("%v bot @ %v", whcb.botPlatform.ID(), whcb.r.Host),
			DataSource:    "bot",
		}
	}
	return gamp.Common{
		DataSource: "bot",
		ClientID:   "", // TODO: DO NOT USE hardcoded value here!
	}
}

func (gac gaContext) GaEvent(category, action string) *gamp.Event { // TODO: remove
	return gamp.NewEvent(category, action, gac.GaCommon())
}

func (gac gaContext) GaEventWithLabel(category, action, label string) *gamp.Event {
	return gamp.NewEventWithLabel(category, action, label, gac.GaCommon())
}

// BotPlatform indicates on which bot platform we process message
func (whcb *WebhookContextBase) BotPlatform() BotPlatform {
	return whcb.botPlatform
}

// GetBotSettings settings of the current bot
func (whcb *WebhookContextBase) GetBotSettings() *BotSettings {
	return whcb.botContext.BotSettings
}

// GetBotCode returns current bot code
func (whcb *WebhookContextBase) GetBotCode() string {
	return whcb.botContext.BotSettings.Code
}

func (whcb *WebhookContextBase) GetBotUserID() string {
	return fmt.Sprintf("%v", whcb.input.GetSender().GetID())
}

// GetBotToken returns current bot API token
func (whcb *WebhookContextBase) GetBotToken() string {
	return whcb.botContext.BotSettings.Token
}

//func (whcb *WebhookContextBase) GetHTTPClient() *http.Client {
//	return whcb.botContext.BotHost.GetHTTPClient(whcb.c)
//}

// HasChatData return true if messages is within botChat
func (whcb *WebhookContextBase) HasChatData() bool {
	return whcb.botChat.Data != nil
}

//func (whcb *WebhookContextBase) SaveAppUser(appUserID int64, appUserEntity BotAppUser) error {
//	return whcb.BotAppUserStore.SaveAppUser(whcb.Context(), appUserID, appUserEntity)
//}

//// SetChatEntity sets botChat data for the context (loaded from DB)
//func (whcb *WebhookContextBase) SetChatEntity(chatData botsfwmodels.BotChatData) {
//	whcb.chatData = chatData
//}

// ChatData returns app entity for the context (loaded from DB)
func (whcb *WebhookContextBase) ChatData() botsfwmodels.BotChatData {
	if whcb.botChat.Data != nil {
		return whcb.botChat.Data
	}
	whcb.isLoadingChatData = true
	defer func() {
		whcb.isLoadingChatData = false
	}()
	//panic("*WebhookContextBase.BotChatData()")
	//log.Debugf(whcb.c, "*WebhookContextBase.BotChatData()")
	chatID, err := whcb.BotChatID()
	if err != nil {
		panic(fmt.Errorf("failed to call whcb.BotChatID(): %w", err))
	}
	if chatID == "" {
		log.Debugf(whcb.c, "whcb.BotChatID() is empty string")
		return nil
	}
	if err := whcb.loadChatEntityBase(); err != nil {
		if dal.IsNotFound(err) {
			botID := whcb.GetBotCode()
			if whcb.recordsFieldsSetter == nil {
				panic("whcb.recordsFieldsSetter == nil")
			}
			sender := whcb.input.GetSender()
			botUserID := fmt.Sprintf("%v", sender.GetID())
			appUserID := whcb.AppUserID()
			webhookChat := whcb.Chat()
			if err = whcb.recordsFieldsSetter.SetBotChatFields(
				whcb.botChat.Data,
				webhookChat,
				botID,
				botUserID,
				appUserID,
				true, // isAccessGranted - TODO: Implement!!!
			); err != nil {
				panic(fmt.Errorf("failed to call whcb.recordsMaker.MakeBotChatDto(): %w", err))
			}
		} else {
			panic(fmt.Errorf("failed to call whcb.loadChatEntityBase(): %w", err))
		}
	}
	return whcb.botChat.Data
}

// getOrCreateBotUserData to be documented
func (whcb *WebhookContextBase) getOrCreateBotUserData() (botsfwmodels.BotUserData, error) {
	if whcb.botUser.Data != nil {
		return whcb.botUser.Data, nil
	}
	c := whcb.Context()
	log.Debugf(c, "getOrCreateBotUserData()")
	whcb.isLoadingUserData = true
	defer func() {
		whcb.isLoadingUserData = false
	}()
	sender := whcb.input.GetSender()
	platformID := whcb.botPlatform.ID()
	botID := whcb.GetBotCode()
	botUserID := fmt.Sprintf("%v", sender.GetID())
	var err error
	whcb.botUser, err = botsdal.GetBotUser(c, whcb.tx, platformID, botID, botUserID, whcb.botContext.BotSettings.Profile.NewBotUserData)
	if err != nil {
		if !dal.IsNotFound(err) {
			log.Infof(c, "Bot user entity not found, creating a new one...")
			appUserID := whcb.AppUserID()
			if err = whcb.recordsFieldsSetter.SetBotUserFields(whcb.botUser.Data, sender, botID, appUserID, botUserID); err != nil {
				log.Errorf(c, "WebhookContextBase.getOrCreateBotUserData(): failed to make bot user DTO: %v", err)
				return whcb.botUser.Data, err
			}
			if err = whcb.SaveBotUser(c); err != nil {
				log.Errorf(c, "WebhookContextBase.getOrCreateBotUserData(): failed to create bot user: %v", err)
				return whcb.botUser.Data, err
			}
			log.Infof(c, "Bot user entity created")

			ga := whcb.gaContext
			if err = ga.Queue(ga.GaEvent("users", "user-created")); err != nil { //TODO: Should be outside
				log.Errorf(c, "Failed to queue GA event: %v", err)
			}

			if err = ga.Queue(ga.GaEventWithLabel("users", "messenger-linked", whcb.botPlatform.ID())); err != nil { // TODO: Should be outside
				log.Errorf(c, "Failed to queue GA event: %v", err)
			}

			if whcb.GetBotSettings().Env == EnvProduction {
				if err = ga.Queue(ga.GaEventWithLabel("bot-users", "bot-user-created", whcb.botPlatform.ID())); err != nil {
					log.Errorf(c, "Failed to queue GA event: %v", err)
				}
			}
		}
		return whcb.botUser.Data, err
	} else {
		log.Infof(c, "Found existing bot user entity")
	}
	return whcb.botUser.Data, err
}

var EnvLocal = "local"           // TODO: Consider adding this to init interface of setting config values
var EnvProduction = "production" // TODO: Consider adding this to init interface of setting config values

func (whcb *WebhookContextBase) loadChatEntityBase() (err error) {
	ctx, cancel := context.WithTimeout(whcb.Context(), time.Second)
	defer cancel()
	if whcb.HasChatData() {
		log.Warningf(ctx, "Duplicate call of func (whc *bot.WebhookContext) _getChat()")
		return nil
	}

	var chatKey = botsfwmodels.ChatKey{
		BotID: whcb.GetBotCode(),
	}
	if chatKey.ChatID, err = whcb.BotChatID(); err != nil {
		return fmt.Errorf("failed to call whcb.BotChatID(): %w", err)
	}

	platformID := whcb.botPlatform.ID()
	whcb.botChat, err = botsdal.GetBotChat(ctx, whcb.tx, platformID,
		whcb.botContext.BotSettings.Code, whcb.botChat.ID, whcb.botContext.BotSettings.Profile.NewBotChatData)
	if err != nil && !dal.IsNotFound(err) {
		return
	}
	if whcb.botChat.Record.Exists() {
		if botUserID := whcb.GetBotUserID(); botUserID != "" {
			chatDataBase := whcb.botChat.Data.Base()
			switch len(chatDataBase.BotUserIDs) {
			case 0:
				chatDataBase.BotUserIDs = []string{botUserID}
			case 1:
				if chatDataBase.BotUserIDs[0] != botUserID {
					// Different bot user ID - should never happen?
					log.Warningf(ctx, "different bot user ID: %s != %s: chatKey=%v", chatKey, chatDataBase.BotUserIDs[0], botUserID)
				}
			default:
				chatDataBase.SetBotUserID(botUserID)
			}
		}
	}
	if err != nil {
		if !dal.IsNotFound(err) {
			return err
		}
		err = nil
		log.Infof(ctx, "BotChat not found, first check for bot user entity...")
		botUser, err := whcb.getOrCreateBotUserData()
		if err != nil {
			return err
		}

		botUserID := fmt.Sprintf("%v", whcb.input.GetSender().GetID())

		isAccessGranted := botUser.IsAccessGranted()
		whChat := whcb.input.Chat()
		appUserID := botUser.GetAppUserID()
		if err = whcb.recordsFieldsSetter.SetBotChatFields(whcb.botChat.Data, whChat, chatKey.BotID, botUserID, appUserID, isAccessGranted); err != nil {
			return err
		}

		if whcb.GetBotSettings().Env == EnvProduction {
			ga := whcb.gaContext
			if err := ga.Queue(ga.GaEventWithLabel("bot-chats", "bot-botChat-created", whcb.botPlatform.ID())); err != nil {
				log.Errorf(ctx, "Failed to queue GA event: %v", err)
			}
		}

	}

	if sender := whcb.input.GetSender(); sender != nil {
		if languageCode := sender.GetLanguage(); languageCode != "" {
			whcb.botChat.Data.AddClientLanguage(languageCode)
		}
	}

	if chatLocale := whcb.botChat.Data.GetPreferredLanguage(); chatLocale != "" && chatLocale != whcb.locale.Code5 {
		if err = whcb.SetLocale(chatLocale); err != nil {
			log.Errorf(ctx, "failed to set locate: %v", err)
		}
	}
	return err
}

// AppUserEntity current app user entity from data storage
func (whcb *WebhookContextBase) AppUserEntity() botsfwmodels.AppUserData {
	return whcb.appUserData
}

// Context for current request
func (whcb *WebhookContextBase) Context() context.Context {
	return whcb.c
}

// SetContext sets current context // TODO: explain why we need this as probably should be in constructor?
func (whcb *WebhookContextBase) SetContext(c context.Context) {
	whcb.c = c
}

// MessageText returns text of a received message
func (whcb *WebhookContextBase) MessageText() string {
	if tm, ok := whcb.Input().(WebhookTextMessage); ok {
		return tm.Text()
	}
	return ""
}

// NewMessageByCode creates new translated message by i18n code
func (whcb *WebhookContextBase) NewMessageByCode(messageCode string, a ...interface{}) (m MessageFromBot) {
	text := whcb.Translate(messageCode)
	text = fmt.Sprintf(text, a...)
	return whcb.NewMessage(text)
}

// NewMessage creates a new text message from bot
func (whcb *WebhookContextBase) NewMessage(text string) (m MessageFromBot) {
	m.Text = text
	return
}

// Locale indicates current language
func (whcb *WebhookContextBase) Locale() i18n.Locale {
	if whcb.locale.Code5 == "" {
		if chatData := whcb.ChatData(); chatData != nil {
			if locale := chatData.GetPreferredLanguage(); locale != "" {
				if err := whcb.SetLocale(locale); err == nil {
					return whcb.locale
				}
			}
		}
		whcb.locale = whcb.botContext.BotSettings.Locale
	}
	return whcb.locale
}

// SetLocale sets current language
func (whcb *WebhookContextBase) SetLocale(code5 string) error {
	if code5 == "" {
		return errors.New("whcb.SetLocate(code5) expects non-empty string")
	}
	if whcb.botAppContext == nil {
		return fmt.Errorf("botAppContext is nil")
	}
	supportedLocales := whcb.botAppContext.SupportedLocales()
	if supportedLocales == nil {
		return fmt.Errorf("supportedLocales is nil")
	}
	locale, err := whcb.botAppContext.GetLocaleByCode5(code5)
	if err != nil {
		return fmt.Errorf(
			"whcb.SetLocate(%s) failed to call supportedLocales.GetLocaleByCode5(%s): %w",
			code5, code5, err)
	}
	whcb.locale = locale
	//log.Debugf(whcb.Context(), "*WebhookContextBase.SetLocale(%v) => Done", code5)
	return nil
}

// CommandText returns a title for a command
func (whcb *WebhookContextBase) CommandText(title, icon string) string {
	if title != "" && !strings.HasPrefix(title, "/") {
		title = whcb.Translate(title)
	}
	return CommandTextNoTrans(title, icon)
}

func (whcb *WebhookContextBase) SaveBotChat(ctx context.Context) error {
	return whcb.tx.Set(ctx, whcb.botChat.Record)
}

func (whcb *WebhookContextBase) SaveBotUser(ctx context.Context) error {
	return whcb.tx.Set(ctx, whcb.botChat.Record)
}
