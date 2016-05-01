package bots

import (
	"net/http"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
)

type UserEntity interface {
	SetPreferredLocale(code5 string) error
	PreferredLocale() string
}

type WebhookContext interface {
	GetLogger() Logger
	BotInputProvider
	Translate(key string) string
	TranslateNoWarning(key string) string

	NewChatEntity() BotChat
	MakeChatEntity() BotChat

	Init(w http.ResponseWriter, r *http.Request) error
	Context() context.Context

	ChatKey() *datastore.Key
	NewChatKey(c context.Context) *datastore.Key
	ChatEntity()  BotChat
	ReplyByBot(m MessageFromBot) error

	CommandTitle(title, icon string) string
	CommandTitleNoTrans(title, icon string) string

	Locale() Locale
	SetLocale(code5 string) error

	NewMessage(text string) MessageFromBot
	NewMessageByCode(messageCode string, a ...interface{}) MessageFromBot

	GetHttpClient() *http.Client
	IsNewerThen(chatEntity BotChat) bool
	UpdateLastProcessed(chatEntity BotChat) error

	GetOrCreateUserEntity() (BotUser, error)
	UserID() int64
	CurrentUserKey() *datastore.Key
	GetUser() (*datastore.Key, UserEntity, error)
	GetOrCreateUser() (*datastore.Key, UserEntity, error)

	ApiUser() BotApiUser
}

type BotInputProvider interface {
	MessageText() string
}

type BotApiUser interface {
	IdAsString() string
	IdAsInt64() int64
	FirstName() string
	LastName() string
}