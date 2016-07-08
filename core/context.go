package bots

import (
	"golang.org/x/net/context"
	"net/http"
)

type WebhookInlineQueryContext interface {
}

type WebhookContext interface {
	GetLogger() Logger
	BotInputProvider
	BotPlatform() BotPlatform

	Init(w http.ResponseWriter, r *http.Request) error
	Context() context.Context

	BotChatID() interface{}

	GetBotCode() string
	GetBotToken() string

	ChatEntity() BotChat

	CommandText(title, icon string) string

	//Locale() Locale
	SetLocale(code5 string) error

	NewMessage(text string) MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot

	GetHttpClient() *http.Client
	UpdateLastProcessed(chatEntity BotChat) error

	AppUserID() int64
	GetAppUser() (AppUser, error)
	SaveAppUser(appUserID int64, appUserEntity AppUser) error

	BotState
	BotChatStore
	BotUserStore
	WebhookInput
	SingleLocaleTranslator

	Responder() WebhookResponder
}

type BotState interface {
	IsNewerThen(chatEntity BotChat) bool
}

type BotInputProvider interface {
	MessageText() string
}

type BotApiUser interface {
	//IdAsString() string
	IdAsInt64() int64
	FirstName() string
	LastName() string
}
