package botsfw

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore"
	"github.com/bots-go-framework/bots-fw/botinput"
	botsfw3 "github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/strongo/i18n"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebhookContextBase provides base implementation of WebhookContext interface
// TODO: Document purpose of a dedicated base struct (e.g. example of usage by developers)
type WebhookContextBase struct {
	//w http.ResponseWriter
	r           *http.Request
	c           context.Context
	appContext  AppContext
	botContext  BotContext // TODO: rename to something strongo
	botPlatform BotPlatform
	input       botinput.InputMessage
	//recordsMaker        botsfwmodels.BotRecordsMaker
	recordsFieldsSetter BotRecordsFieldsSetter

	whAnalytics WebhookAnalytics

	getIsInGroup func() (bool, error)

	getLocaleAndChatID func() (locale, chatID string, err error) // TODO: Document why we need to pass context. Is it to support transactions?

	locale i18n.Locale

	chatID            string
	linkedIdentity    *botsfwstore.LinkedIdentity
	isLoadingChatData bool

	appUserID   string
	appUserData botsfwmodels.AppUserData

	translator
	//Locales    strongoapp.LocalesProvider

	store botsfwstore.StateStore
}

func (whcb *WebhookContextBase) Analytics() WebhookAnalytics {
	return whcb.whAnalytics
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
	whcb.chatID = chatID
}

// LogRequest logs request data to logging system
func (whcb *WebhookContextBase) LogRequest() {
	whcb.input.LogRequest()
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
	if whcb.chatID != "" {
		return whcb.chatID, nil
	}
	//log.Debugf(whcb.c, "*WebhookContextBase.BotChatID()")

	input := whcb.Input()
	if botChatID, err = input.BotChatID(); err != nil {
		return
	} else if botChatID != "" {
		whcb.SetChatID(botChatID)
		return whcb.chatID, nil
	}
	if whcb.getLocaleAndChatID != nil {
		if _, botChatID, err = whcb.getLocaleAndChatID(); err != nil {
			return
		}
		if botChatID != "" {
			whcb.SetChatID(botChatID)
			return whcb.chatID, nil
		}
	}
	switch input := input.(type) {
	case botinput.CallbackQuery:
		data := input.GetData()
		if strings.Contains(data, "botChat=") {
			values, err := url.ParseQuery(data)
			if err != nil {
				return "", fmt.Errorf("failed to GetData() from webhookInput.InputCallbackQuery(): %w", err)
			}
			chatID := values.Get("botChat")
			whcb.SetChatID(chatID)
		}
	case botinput.InlineQuery:
		// pass
	case botinput.ChosenInlineResult:
		// pass
	default:
		whcb.LogRequest()
		log.Debugf(whcb.c, "BotChatID(): *.WebhookContextBaseBotChatID(): Unhandled input type: %T", input)
	}

	return whcb.chatID, nil
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

func (whcb *WebhookContextBase) SetUser(id string, data botsfwmodels.AppUserData) {
	whcb.appUserID = id
	whcb.appUserData = data
}

// AppUserID returns the application user linked to the current bot identity.
func (whcb *WebhookContextBase) AppUserID() string {
	if whcb.appUserID != "" {
		return whcb.appUserID
	}
	if whcb.isLoadingChatData {
		return ""
	}
	if whcb.ChatData() != nil && whcb.linkedIdentity != nil {
		whcb.SetUser(whcb.linkedIdentity.AppUser.ID, whcb.linkedIdentity.AppUser.Data)
	}
	if whcb.appUserID == "" {
		chatID, err := whcb.BotChatID()
		if err != nil {
			panic(fmt.Errorf("resolve bot chat before linking application user: %w", err))
		}
		if chatID == "" {
			whcb.linkIdentityWithoutChat()
		}
	}
	return whcb.appUserID
}

// linkIdentityWithoutChat resolves the platform user for chat-independent
// updates such as Telegram inline queries. Identity linking belongs to the
// platform user, not to a bot chat; a chat record is therefore optional.
func (whcb *WebhookContextBase) linkIdentityWithoutChat() {
	if whcb.linkedIdentity != nil {
		whcb.SetUser(
			whcb.linkedIdentity.AppUser.ID,
			whcb.linkedIdentity.AppUser.Data,
		)
		return
	}
	if whcb.recordsFieldsSetter == nil {
		panic("whcb.recordsFieldsSetter == nil")
	}
	sender := whcb.input.GetSender()
	identity := whcb.identity("")
	request := botsfwstore.LinkRequest{
		Identity:             identity,
		ReadPlatformUserData: whcb.botContext.BotSettings.Profile.NewPlatformUserData,
		NewPlatformUserData: func(appUserID string) (botsfwmodels.PlatformUserData, error) {
			data := whcb.botContext.BotSettings.Profile.NewPlatformUserData()
			if data == nil {
				return nil, errors.New("bot profile returned nil platform-user data")
			}
			if err := whcb.recordsFieldsSetter.SetBotUserFields(
				data,
				sender,
				identity.BotID,
				identity.BotUserID,
				appUserID,
			); err != nil {
				return nil, fmt.Errorf("set platform-user fields: %w", err)
			}
			return data, nil
		},
	}
	if err := request.Validate(); err != nil {
		panic(fmt.Errorf("invalid chatless identity-link request: %w", err))
	}
	linked, err := whcb.store.EnsureLinked(whcb.Context(), request)
	if err != nil {
		panic(fmt.Errorf("ensure linked chatless bot identity: %w", err))
	}
	whcb.linkedIdentity = &linked
	whcb.SetUser(linked.AppUser.ID, linked.AppUser.Data)
}

func (whcb *WebhookContextBase) identity(chatID string) botsfwstore.Identity {
	sender := whcb.input.GetSender()
	identity := botsfwstore.Identity{
		PlatformID: whcb.botPlatform.ID(),
		BotID:      whcb.GetBotCode(),
		BotUserID:  whcb.GetBotUserID(),
		ChatID:     chatID,
	}
	if sender != nil {
		identity.FirstName = sender.GetFirstName()
		identity.LastName = sender.GetLastName()
		identity.Username = sender.GetUserName()
		identity.LanguageCode = sender.GetLanguage()
	}
	return identity
}

func (whcb *WebhookContextBase) GetBotUser() (botsfwstore.PlatformUser, error) {
	if whcb.linkedIdentity != nil && whcb.linkedIdentity.PlatformUser.Data != nil {
		return whcb.linkedIdentity.PlatformUser, nil
	}
	if whcb.ChatData() != nil && whcb.linkedIdentity != nil {
		return whcb.linkedIdentity.PlatformUser, nil
	}
	return whcb.store.PlatformUser(whcb.Context(), whcb.identity(""), whcb.botContext.BotSettings.Profile.NewPlatformUserData)
}

func (whcb *WebhookContextBase) SetBotUserAccessGranted(value bool) error {
	updated, err := whcb.store.SetPlatformUserAccessGranted(whcb.Context(), whcb.identity(""), whcb.botContext.BotSettings.Profile.NewPlatformUserData, value)
	if err != nil {
		return err
	}
	if whcb.linkedIdentity != nil {
		whcb.linkedIdentity.PlatformUser = updated
	}
	return nil
}

// GetAppUser loads information about the current app user through the state-store port.
func (whcb *WebhookContextBase) GetAppUser() (botsfwmodels.AppUserData, error) {
	return whcb.AppUserData()
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
	if args.Store == nil {
		panic("args.Store == nil")
	}
	c := args.BotContext.BotHost.Context(args.HttpRequest)
	whcb = &WebhookContextBase{
		r:     args.HttpRequest,
		c:     c,
		store: args.Store,
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
	whcb.whAnalytics = webhookAnalytics{whcb: whcb}
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

func (whcb *WebhookContextBase) GetTranslator(locale string) i18n.SingleLocaleTranslator {
	return translator{
		localeCode5: func() string {
			return locale
		},
		Translator: whcb.appContext.GetTranslator(whcb.c),
	}
}

// Input returns webhook input
func (whcb *WebhookContextBase) Input() botinput.InputMessage {
	return whcb.input
}

// Chat returns webhook botChat
func (whcb *WebhookContextBase) Chat() botinput.Chat { // TODO: remove
	return whcb.input.Chat()
}

// GetRecipient returns receiver of the message
func (whcb *WebhookContextBase) GetRecipient() botinput.Recipient { // TODO: remove
	return whcb.input.GetRecipient()
}

// GetSender returns sender of the message
//func (whcb *WebhookContextBase) GetSender() botinput.User { // TODO: remove
//	return whcb.input.GetSender()
//}

// GetTime returns time of the message
func (whcb *WebhookContextBase) GetTime() time.Time { // TODO: remove
	return whcb.input.GetTime()
}

// InputType returns input type
func (whcb *WebhookContextBase) InputType() botinput.Type { // TODO: remove
	return whcb.input.InputType()
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
	return whcb.linkedIdentity != nil && whcb.linkedIdentity.ChatData != nil
}

// ChatData returns current bot chat state, creating and linking the identity
// through the injected store when the platform supplied a chat ID.
func (whcb *WebhookContextBase) ChatData() botsfwmodels.BotChatData {
	if whcb.HasChatData() {
		return whcb.linkedIdentity.ChatData
	}
	whcb.isLoadingChatData = true
	defer func() { whcb.isLoadingChatData = false }()
	chatID, err := whcb.BotChatID()
	if err != nil {
		panic(fmt.Errorf("failed to call whcb.BotChatID(): %w", err))
	}
	if chatID == "" {
		log.Debugf(whcb.c, "whcb.BotChatID() is empty string")
		return nil
	}
	if whcb.recordsFieldsSetter == nil {
		panic("whcb.recordsFieldsSetter == nil")
	}
	sender := whcb.input.GetSender()
	identity := whcb.identity(chatID)
	request := botsfwstore.LinkRequest{
		Identity:             identity,
		ReadPlatformUserData: whcb.botContext.BotSettings.Profile.NewPlatformUserData,
		NewPlatformUserData: func(appUserID string) (botsfwmodels.PlatformUserData, error) {
			data := whcb.botContext.BotSettings.Profile.NewPlatformUserData()
			if data == nil {
				return nil, errors.New("bot profile returned nil platform-user data")
			}
			if err := whcb.recordsFieldsSetter.SetBotUserFields(data, sender, identity.BotID, identity.BotUserID, appUserID); err != nil {
				return nil, fmt.Errorf("set platform-user fields: %w", err)
			}
			return data, nil
		},
		NewChatData: func(appUserID string, accessGranted bool) (botsfwmodels.BotChatData, error) {
			data := whcb.botContext.BotSettings.Profile.NewBotChatData()
			if data == nil {
				return nil, errors.New("bot profile returned nil chat data")
			}
			if err := whcb.recordsFieldsSetter.SetBotChatFields(data, whcb.Chat(), identity.BotID, identity.BotUserID, appUserID, accessGranted); err != nil {
				return nil, fmt.Errorf("set chat fields: %w", err)
			}
			return data, nil
		},
	}
	if err := request.Validate(); err != nil {
		panic(fmt.Errorf("invalid identity-link request: %w", err))
	}
	linked, err := whcb.store.EnsureLinked(whcb.Context(), request)
	if err != nil {
		panic(fmt.Errorf("ensure linked bot identity: %w", err))
	}
	whcb.linkedIdentity = &linked
	whcb.SetUser(linked.AppUser.ID, linked.AppUser.Data)
	if linked.ChatData == nil {
		return nil
	}
	if sender != nil && sender.GetLanguage() != "" {
		linked.ChatData.AddClientLanguage(sender.GetLanguage())
	}
	if chatLocale := linked.ChatData.GetPreferredLanguage(); chatLocale != "" && chatLocale != whcb.locale.Code5 {
		if err := whcb.SetLocale(chatLocale); err != nil {
			log.Errorf(whcb.Context(), "failed to set locale: %v", err)
		}
	}
	return linked.ChatData
}

var EnvLocal = "local"           // TODO: Consider adding this to init interface of setting config values
var EnvProduction = "production" // TODO: Consider adding this to init interface of setting config values

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
	if tm, ok := whcb.Input().(botinput.TextMessage); ok {
		return tm.Text()
	}
	return ""
}

// NewMessageByCode creates new translated message by i18n code
func (whcb *WebhookContextBase) NewMessageByCode(messageCode string, a ...interface{}) (m botsfw3.MessageFromBot) {
	text := whcb.Translate(messageCode)
	text = fmt.Sprintf(text, a...)
	return whcb.NewMessage(text)
}

// NewMessage creates a new text message from bot
func (whcb *WebhookContextBase) NewMessage(text string) (m botsfw3.MessageFromBot) {
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
	chatData := whcb.ChatData()
	if chatData == nil {
		return nil
	}
	chatID, err := whcb.BotChatID()
	if err != nil {
		return err
	}
	return whcb.store.SaveChat(whcb.Context(), whcb.identity(chatID), chatData)
}

func (whcb *WebhookContextBase) AppUserData() (appUserData botsfwmodels.AppUserData, err error) {
	appUserID := whcb.AppUserID()
	if appUserID == "" {
		return nil, fmt.Errorf("%w: AppUserID() is empty", botsfwstore.ErrNotFound)
	}
	if whcb.appUserData != nil {
		return whcb.appUserData, nil
	}
	appUser, err := whcb.store.AppUser(whcb.Context(), whcb.GetBotCode(), appUserID)
	if err != nil {
		return
	}
	whcb.SetUser(appUser.ID, appUser.Data)
	return appUser.Data, nil
}
