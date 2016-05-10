package bots

import (
	"golang.org/x/net/context"
	"net/http"
)

type WebhookContext interface {
	GetLogger() Logger
	BotInputProvider
	Translate(key string) string
	TranslateNoWarning(key string) string

	Init(w http.ResponseWriter, r *http.Request) error
	Context() context.Context

	BotChatID() interface{}

	ChatEntity() BotChat
	ReplyByBot(w http.ResponseWriter, m MessageFromBot) error

	CommandTitle(title, icon string) string

	Locale() Locale
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
