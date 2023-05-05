package botsfw

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/i18n"
	"net/http"

	"context"
	"github.com/strongo/app"
	"github.com/strongo/gamp"
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
	dal.TransactionCoordinator
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

	ChatEntity() BotChat

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

	UpdateLastProcessed(chatEntity BotChat) error

	AppUserID() string

	GetAppUser() (BotAppUser, error)
	//SaveAppUser(appUserID int64, appUserEntity BotAppUser) error

	BotState
	BotChatStore
	BotUserStore
	WebhookInput // TODO: Should be removed!!!
	i18n.SingleLocaleTranslator

	Responder() WebhookResponder
}

// BotState provides state of the bot (TODO: document how is used)
type BotState interface {
	IsNewerThen(chatEntity BotChat) bool
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
