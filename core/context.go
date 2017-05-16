package bots

import (
	"github.com/strongo/app"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"net/http"
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
	Environment() strongo.Environment
	BotInputProvider
	BotPlatform() BotPlatform

	Init(w http.ResponseWriter, r *http.Request) error
	Context() context.Context

	ExecutionContext() strongo.ExecutionContext
	BotAppContext() BotAppContext

	BotChatID() string

	GetBotCode() string
	GetBotToken() string
	GetBotSettings() BotSettings

	ChatEntity() BotChat

	CommandText(title, icon string) string

	//Locale() strongo.Locale
	SetLocale(code5 string) error

	NewMessage(text string) MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot
	NewEditCallbackMessage(messageText string) MessageFromBot
	//NewEditCallbackMessageKeyboard(kbMarkup tgbotapi.InlineKeyboardMarkup) MessageFromBot

	GetHttpClient() *http.Client
	UpdateLastProcessed(chatEntity BotChat) error

	AppUserIntID() int64
	GetAppUser() (BotAppUser, error)
	SaveAppUser(appUserID int64, appUserEntity BotAppUser) error

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
