package botsfw

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsdal"
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
	r           *http.Request
	c           context.Context
	appContext  AppContext
	botContext  BotContext // TODO: rename to something strongo
	botPlatform BotPlatform
	input       botinput.WebhookInput
	//recordsMaker        botsfwmodels.BotRecordsMaker
	recordsFieldsSetter BotRecordsFieldsSetter

	getIsInGroup func() (bool, error)

	getLocaleAndChatID func() (locale, chatID string, err error) // TODO: Document why we need to pass context. Is it to support transactions?

	locale i18n.Locale

	// At the moment, there is no reason to expose botChat record publicly
	// If there is some it should be documented with a use case
	botChat      record.DataWithID[string, botsfwmodels.BotChatData]
	platformUser botsdal.BotUser // Telegram user ID is an integer, but we keep it as a string for consistency & simplicity.

	isLoadingChatData         bool // TODO: This smells bad. Needs refactoring?
	isLoadingPlatformUserData bool // TODO: This smells bad. Needs refactoring?

	//
	appUserID   string
	appUserData botsfwmodels.AppUserData

	translator
	//Locales    strongoapp.LocalesProvider

	//dal botsfwdal.DataAccess
	db dal.DB
	//tx dal.ReadwriteTransaction

	gaContext gaContext
}

func (whcb *WebhookContextBase) DB() dal.DB {
	return whcb.db
}

// Tx returns a transaction that is used to read/write botChat & bot user data
//func (whcb *WebhookContextBase) Tx() dal.ReadwriteTransaction {
//	return whcb.tx
//}

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
	case botinput.WebhookCallbackQuery:
		data := input.GetData()
		if strings.Contains(data, "botChat=") {
			values, err := url.ParseQuery(data)
			if err != nil {
				return "", fmt.Errorf("failed to GetData() from webhookInput.InputCallbackQuery(): %w", err)
			}
			chatID := values.Get("botChat")
			whcb.SetChatID(chatID)
		}
	case botinput.WebhookInlineQuery:
		// pass
	case botinput.WebhookChosenInlineResult:
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
	if whcb.appUserID == "" && !whcb.isLoadingChatData {
		if chatData := whcb.ChatData(); chatData != nil {
			whcb.appUserID = chatData.GetAppUserID()
		}
	}
	if whcb.platformUser.Data == nil {
		var err error
		if err = whcb.getPlatformUserRecord(whcb.db); err != nil {
			if !dal.IsNotFound(err) {
				panic(fmt.Errorf("failed to get bot user entity: %w", err))
			}
		}
	}
	if whcb.platformUser.Data != nil {
		whcb.appUserID = whcb.platformUser.Data.GetAppUserID()
	}
	return whcb.appUserID
	//if appUserID == "" && !whcb.isLoadingPlatformUserData {
	//	if platformUser, err := whcb.getOrCreatePlatformUserRecord(); err != nil {
	//		if !dal.IsNotFound(err) {
	//			panic(fmt.Errorf("failed to get bot user entity: %w", err))
	//		}
	//	} else {
	//		appUserID = platformUser.GetAppUserID()
	//	}
	//}
}

func (whcb *WebhookContextBase) GetBotUserForUpdate(ctx context.Context, tx dal.ReadwriteTransaction) (botUser botsdal.BotUser, err error) {
	err = whcb.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		botUser, err = whcb.getBotUser(ctx, tx)
		return
	})
	return
}

func (whcb *WebhookContextBase) getBotUser(ctx context.Context, tx dal.Getter) (botUser botsdal.BotUser, err error) {
	if whcb.platformUser.Data != nil {
		return whcb.platformUser, nil
	}
	platformID := whcb.BotPlatform().ID()
	botUserID := whcb.GetBotUserID()
	whcb.platformUser, err = botsdal.GetPlatformUser(ctx, whcb.db, platformID, botUserID, whcb.botContext.BotSettings.Profile.NewPlatformUserData())
	return whcb.platformUser, err
}

func (whcb *WebhookContextBase) GetBotUser() (botUser botsdal.BotUser, err error) {
	return whcb.getBotUser(whcb.c, whcb.db)
}

// GetAppUser loads information about current app user from persistent storage
func (whcb *WebhookContextBase) GetAppUser() (botsfwmodels.AppUserData, error) { // TODO: Can/should this be cached?
	appUserID := whcb.AppUserID()
	appUser, err := whcb.BotContext().BotSettings.GetAppUserByID(whcb.c, whcb.db, appUserID)
	if err != nil {
		return nil, err
	}
	return appUser.Data, err
}

// ExecutionContext returns an execution context for strongo app
func (whcb *WebhookContextBase) ExecutionContext() ExecutionContext {
	return whcb
}

// AppContext returns bot app context
func (whcb *WebhookContextBase) AppContext() AppContext {
	return whcb.appContext
}

// IsInGroup signals if the bot request is send within group botChat
func (whcb *WebhookContextBase) IsInGroup() (bool, error) {
	return whcb.getIsInGroup()
}

// NewWebhookContextBase creates base bot context
func NewWebhookContextBase(
	args CreateWebhookContextArgs,
	botPlatform BotPlatform,
	recordsFieldsSetter BotRecordsFieldsSetter, // TODO: Should it be a member of BotContext?
	getIsInGroup func() (bool, error),
	getLocaleAndChatID func(c context.Context) (locale, chatID string, err error),
) (whcb *WebhookContextBase, err error) {
	if args.HttpRequest == nil {
		panic("args.HttpRequest == nil")
	}
	c := args.BotContext.BotHost.Context(args.HttpRequest)
	whcb = &WebhookContextBase{
		r:  args.HttpRequest,
		c:  c,
		db: args.Db,
		//tx: args.Tx,
		getLocaleAndChatID: func() (locale, chatID string, err error) {
			return getLocaleAndChatID(c)
		},
		appContext:   args.AppContext,
		botPlatform:  botPlatform,
		botContext:   args.BotContext,
		input:        args.WebhookInput,
		getIsInGroup: getIsInGroup,
		//dal:                 botCoreStores,
		recordsFieldsSetter: recordsFieldsSetter,
	}
	whcb.gaContext = gaContext{
		whcb:          whcb,
		gaMeasurement: args.GaMeasurement,
	}
	// TODO: make sure we do not fail here for non group chats
	//var isInGroup bool
	//if isInGroup, err = getIsInGroup(); err != nil {
	//	return
	//} else if isInGroup && whcb.getLocaleAndChatID != nil {
	//	var locale, chatID string
	//	if locale, chatID, err = whcb.getLocaleAndChatID(); err != nil {
	//		err = fmt.Errorf("failed in whcb.getLocaleAndChatID(): %w", err)
	//		return
	//	} else {
	//		if chatID != "" {
	//			whcb.SetChatID(chatID)
	//		}
	//		if locale != "" {
	//			if err := whcb.SetLocale(locale); err != nil {
	//				log.Errorf(c, "Failed to set Locale: %v", err)
	//			}
	//		}
	//	}
	//}
	whcb.translator = translator{
		localeCode5: func() string {
			return whcb.locale.Code5
		},
		Translator: args.AppContext.GetTranslator(whcb.c),
	}
	return
}

// Input returns webhook input
func (whcb *WebhookContextBase) Input() botinput.WebhookInput {
	return whcb.input
}

// Chat returns webhook botChat
func (whcb *WebhookContextBase) Chat() botinput.WebhookChat { // TODO: remove
	return whcb.input.Chat()
}

// GetRecipient returns receiver of the message
func (whcb *WebhookContextBase) GetRecipient() botinput.WebhookRecipient { // TODO: remove
	return whcb.input.GetRecipient()
}

// GetSender returns sender of the message
//func (whcb *WebhookContextBase) GetSender() botinput.WebhookUser { // TODO: remove
//	return whcb.input.GetSender()
//}

// GetTime returns time of the message
func (whcb *WebhookContextBase) GetTime() time.Time { // TODO: remove
	return whcb.input.GetTime()
}

// InputType returns input type
func (whcb *WebhookContextBase) InputType() botinput.WebhookInputType { // TODO: remove
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
func (gac gaContext) GaCommon() (result gamp.Common) {
	whcb := gac.whcb
	if whcb.botChat.Record != nil && whcb.botChat.Record.Exists() {
		result.UserID = whcb.botChat.Data.GetAppUserID()
		result.UserLanguage = strings.ToLower(whcb.botChat.Data.GetPreferredLanguage())
		platformID := whcb.botPlatform.ID()
		result.ApplicationID = fmt.Sprintf("bot.%v.%v", platformID, whcb.GetBotCode())
		result.UserAgent = fmt.Sprintf("%v bot @ %v", platformID, whcb.r.Host)
		result.DataSource = "bot"
		return
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
	if err = whcb.loadChatEntityBase(); err != nil {
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

func (whcb *WebhookContextBase) getPlatformUserRecord(tx dal.ReadSession) (err error) {
	if whcb.platformUser.Data != nil {
		return nil
	}
	platformID := whcb.botPlatform.ID()
	sender := whcb.input.GetSender()
	ctx := whcb.Context()

	whcb.platformUser.ID = fmt.Sprintf("%v", sender.GetID())
	whcb.platformUser.Data = whcb.botContext.BotSettings.Profile.NewPlatformUserData()
	if whcb.platformUser, err = botsdal.GetPlatformUser(ctx, tx, platformID, whcb.platformUser.ID, whcb.platformUser.Data); err != nil {
		return
	}
	return
}

func (whcb *WebhookContextBase) createPlatformUserRecord(tx dal.ReadwriteTransaction) (err error) {
	if whcb.platformUser.Data != nil {
		return nil
	}
	//platformID := whcb.botPlatform.ID()
	botID := whcb.GetBotCode()
	sender := whcb.input.GetSender()
	botUserID := fmt.Sprintf("%v", sender.GetID())
	ctx := whcb.Context()

	if err = whcb.recordsFieldsSetter.SetBotUserFields(whcb.platformUser.Data, sender, botID, botUserID, botUserID); err != nil {
		log.Errorf(ctx, "WebhookContextBase.getOrCreatePlatformUserRecord(): failed to make bot user DTO: %v", err)
		return err
	}
	if err = tx.Set(ctx, whcb.platformUser.Record); err != nil {
		log.Errorf(ctx, "WebhookContextBase.getOrCreatePlatformUserRecord(): failed to create bot user: %v", err)
		return err
	}
	log.Infof(ctx, "Bot user entity created")

	{ // Log analytics
		ga := whcb.gaContext
		if err = ga.Queue(ga.GaEvent("users", "user-created")); err != nil { //TODO: Should be outside
			log.Errorf(ctx, "Failed to queue GA event: %v", err)
			err = nil
			return
		}
		if err = ga.Queue(ga.GaEventWithLabel("users", "messenger-linked", whcb.botPlatform.ID())); err != nil { // TODO: Should be outside
			log.Errorf(ctx, "Failed to queue GA event: %v", err)
			err = nil
			return
		}

		if whcb.GetBotSettings().Env == EnvProduction {
			if err = ga.Queue(ga.GaEventWithLabel("bot-users", "bot-user-created", whcb.botPlatform.ID())); err != nil {
				log.Errorf(ctx, "Failed to queue GA event: %v", err)
				err = nil
				return
			}
		}
	}

	return
}

// getOrCreatePlatformUserRecord to be documented
func (whcb *WebhookContextBase) getOrCreatePlatformUserRecord() (botUser botsdal.BotUser, err error) {
	if whcb.platformUser.Data != nil {
		return whcb.platformUser, nil
	}
	ctx := whcb.Context()
	log.Debugf(ctx, "getOrCreatePlatformUserRecord()")
	whcb.isLoadingPlatformUserData = true
	defer func() {
		whcb.isLoadingPlatformUserData = false
	}()

	if err = whcb.getPlatformUserRecord(whcb.db); err != nil {
		if !dal.IsNotFound(err) {
			return whcb.platformUser, err
		} else {
			log.Debugf(ctx, "Bot user entity not found, creating a new one...")
			if err = whcb.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				if err = whcb.createPlatformUserRecord(tx); err != nil {
					return fmt.Errorf("failed to create platform user record: %w", err)
				}
				return nil
			}); err != nil {
				return whcb.platformUser, err
			}
		}

		return whcb.platformUser, err
	} else {
		log.Infof(ctx, "Found existing bot user entity")
	}
	return whcb.platformUser, err
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
	db := whcb.DB()
	whcb.botChat, err = botsdal.GetBotChat(ctx, db, platformID,
		whcb.botContext.BotSettings.Code, whcb.botChat.ID, whcb.botContext.BotSettings.Profile.NewBotChatData)
	if err != nil && !dal.IsNotFound(err) {
		return fmt.Errorf("failed to get bot char record: %w", err)
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
	if dal.IsNotFound(err) {
		log.Infof(ctx, "BotChat not found, first check for bot user entity...")
		var botUser botsdal.BotUser

		if botUser, err = whcb.getOrCreatePlatformUserRecord(); err != nil {
			return err
		}

		isAccessGranted := botUser.Data.IsAccessGranted()
		whChat := whcb.input.Chat()
		appUserID := botUser.Data.GetAppUserID()
		if err = whcb.recordsFieldsSetter.SetBotChatFields(whcb.botChat.Data, whChat, chatKey.BotID, botUser.ID, appUserID, isAccessGranted); err != nil {
			return err
		}

		if whcb.GetBotSettings().Env == EnvProduction {
			ga := whcb.gaContext
			if err = ga.Queue(ga.GaEventWithLabel("bot-chats", "bot-botChat-created", whcb.botPlatform.ID())); err != nil {
				log.Errorf(ctx, "Failed to queue GA event: %v", err)
				err = nil
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
			err = nil
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
	if tm, ok := whcb.Input().(botinput.WebhookTextMessage); ok {
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
	if whcb.appContext == nil {
		return fmt.Errorf("appContext is nil")
	}
	supportedLocales := whcb.appContext.SupportedLocales()
	if supportedLocales == nil {
		return fmt.Errorf("supportedLocales is nil")
	}
	locale, err := whcb.appContext.GetLocaleByCode5(code5)
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

func (whcb *WebhookContextBase) SaveBotChat() error {
	ctx := whcb.Context()
	// It is dangerous to allow user to pass context to this func as if it's a transactional context it might lead to deadlock
	return whcb.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(whcb.c, whcb.botChat.Record)
	})
}

func (whcb *WebhookContextBase) SaveBotUser(ctx context.Context) error {
	return whcb.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return errors.New("func SaveBotUser is not implemented yet")
		//return tx.Set(ctx, whcb.platformUser.Record)
	})
}

func (whcb *WebhookContextBase) AppUserData() (appUserData botsfwmodels.AppUserData, err error) {
	appUserID := whcb.AppUserID()
	if appUserID == "" {
		return nil, fmt.Errorf("%w: AppUserID() is empty", dal.ErrRecordNotFound)
	}
	ctx := whcb.Context()
	botContext := whcb.BotContext()
	var appUser record.DataWithID[string, botsfwmodels.AppUserData]
	if appUser, err = botContext.BotSettings.GetAppUserByID(ctx, whcb.db, appUserID); err != nil {
		return
	}
	return appUser.Data, err
}
