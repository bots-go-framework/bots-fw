package bots

import (
	"github.com/strongo/app"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"net/http"
	"github.com/strongo/db"
)

type WebhookInlineQueryContext interface {
}

type GaContext interface {
	GaMeasurement() *measurement.BufferedSender
	GaCommon() measurement.Common
	GaEvent(category, action string) measurement.Event
	GaEventWithLabel(category, action, label string) measurement.Event
}

type WebhookContext interface {
	GaContext
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

	MustBotChatID() (string)

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

	GetHttpClient() *http.Client
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

type BotState interface {
	IsNewerThen(chatEntity BotChat) bool
}

type BotInputProvider interface {
	Input() WebhookInput
}

type BotApiUser interface {
	//IdAsString() string
	//IdAsInt64() int64
	FirstName() string
	LastName() string
}
