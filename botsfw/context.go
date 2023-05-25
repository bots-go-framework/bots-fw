package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwdal"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/app"
	"github.com/strongo/gamp"
	"github.com/strongo/i18n"
	"net/http"
)

// WebhookInlineQueryContext provides context for inline query (TODO: check & document)
type WebhookInlineQueryContext interface {
}

// GaQueuer queues messages for sending to Google Analytics
type GaQueuer interface { // TODO: can be unexported?
	Queue(message gamp.Message) error
}

// GaContext provides context to Google Analytics
type GaContext interface {
	GaQueuer
	// Flush() error
	GaCommon() gamp.Common
	GaEvent(category, action string) *gamp.Event
	GaEventWithLabel(category, action, label string) *gamp.Event
}

// WebhookContext provides context for current request from user to bot
type WebhookContext interface { // TODO: Make interface much smaller?
	GA() GaContext
	//dal.TransactionCoordinator
	Environment() strongo.Environment
	BotInputProvider
	BotPlatform() BotPlatform

	Request() *http.Request

	//Init(w http.ResponseWriter, r *http.Request) error

	// Context return context
	Context() context.Context

	// SetContext sets context
	SetContext(c context.Context)

	ExecutionContext() strongo.ExecutionContext
	BotAppContext() BotAppContext
	BotContext() BotContext

	MustBotChatID() string

	GetBotCode() string
	GetBotToken() string
	GetBotSettings() BotSettings

	ChatData() botsfwmodels.ChatData // Formerly ChatEntity()
	//ChatKey() botsfwmodels.ChatKey -- commented out as we have it in ChatData but might consider to have it here as well

	// IsInGroup indicates if message was received in a group chat
	IsInGroup() bool

	// CommandText TODO: needs to be documented
	CommandText(title, icon string) string

	//Locale() strongo.ByLocale

	// SetLocale sets locale for current session
	SetLocale(code5 string) error

	NewMessage(text string) MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot
	NewEditMessage(text string, format MessageFormat) (MessageFromBot, error)
	//NewEditMessageKeyboard(kbMarkup tgbotapi.InlineKeyboardMarkup) MessageFromBot

	UpdateLastProcessed(chatEntity botsfwmodels.ChatData) error

	AppUserID() string

	// Deprecated: use AppUserID() instead
	AppUserInt64ID() int64

	AppUserData() (botsfwmodels.AppUserData, error)
	//SaveAppUser(appUserID int64, appUserEntity BotAppUser) error

	BotState

	Store() botsfwdal.DataAccess

	//RecordsMaker() botsfwmodels.BotRecordsMaker

	// RecordsFieldsSetter returns a helper that sets fields of bot related records
	RecordsFieldsSetter() BotRecordsFieldsSetter

	WebhookInput // TODO: Should be removed!!!
	i18n.SingleLocaleTranslator

	Responder() WebhookResponder
}

// BotState provides state of the bot (TODO: document how is used)
type BotState interface {
	IsNewerThen(chatEntity botsfwmodels.ChatData) bool
}

// BotInputProvider provides an input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
type BotInputProvider interface {
	// Input returns a webhook input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
	Input() WebhookInput
}

// BotAPIUser provides info about current bot user
type BotAPIUser interface {
	// FirstName returns user's first name
	FirstName() string

	// LastName returns user's last name
	LastName() string

	//IdAsString() string
	//IdAsInt64() int64

}
