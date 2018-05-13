package bots

import (
	"net/http"

	"context"
	"github.com/strongo/app"
	"github.com/strongo/db"
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
type WebhookContext interface {
	// TODO: Make interface smaller?
	GA() GaContext
	db.TransactionCoordinator
	Environment() strongo.Environment
	BotInputProvider
	BotPlatform() BotPlatform

	Request() *http.Request

	//Init(w http.ResponseWriter, r *http.Request) error
	Context() context.Context
	SetContext(c context.Context)

	ExecutionContext() strongo.ExecutionContext
	BotAppContext() BotAppContext

	MustBotChatID() string

	GetBotCode() string
	GetBotToken() string
	GetBotSettings() BotSettings

	ChatEntity() BotChat

	IsInGroup() bool

	CommandText(title, icon string) string

	//Locale() strongo.ByLocale
	SetLocale(code5 string) error

	NewMessage(text string) MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot
	NewEditMessage(text string, format MessageFormat) (MessageFromBot, error)
	//NewEditMessageKeyboard(kbMarkup tgbotapi.InlineKeyboardMarkup) MessageFromBot

	UpdateLastProcessed(chatEntity BotChat) error

	AppUserIntID() int64
	AppUserStrID() string

	GetAppUser() (BotAppUser, error)
	//SaveAppUser(appUserID int64, appUserEntity BotAppUser) error

	BotState
	BotChatStore
	BotUserStore
	WebhookInput
	strongo.SingleLocaleTranslator

	Responder() WebhookResponder
}

// BotState provides state of the bot (TODO: document how is used)
type BotState interface {
	IsNewerThen(chatEntity BotChat) bool
}

// BotInputProvider provides an input from a specific bot interface (Telegram, FB Messenger, Viber, etc.)
type BotInputProvider interface {
	Input() WebhookInput
}

// BotAPIUser provides info about current bot user
type BotAPIUser interface {
	//IdAsString() string
	//IdAsInt64() int64
	FirstName() string
	LastName() string
}
