package bots

import (
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"time"
)

type BotChat interface {
	GetBotID() string
	SetBotID(botID string)

	GetAppUserIntID() int64
	SetAppUserIntID(id int64)

	AddClientLanguage(languageCode string) (changed bool)

	//GetBotUserIntID() int
	//GetBotUserStringID() string
	SetBotUserID(id interface{})
	SetIsGroupChat(bool)

	IsAccessGranted() bool
	IsGroupChat() bool
	SetAccessGranted(value bool) bool

	GetPreferredLanguage() string
	SetPreferredLanguage(value string)

	SetDtUpdated(time time.Time)
	SetDtLastInteraction(time time.Time)

	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(path string)
	IsAwaitingReplyTo(code string) bool
	AddWizardParam(key, value string)
	GetWizardParam(key string) string
	PopStepsFromAwaitingReplyUpToSpecificParent(code string)
	PushStepToAwaitingReplyTo(code string)
	GetGaClientID() uuid.UUID
}

type BotChatStore interface {
	GetBotChatEntityByID(c context.Context, botID, botChatID string) (BotChat, error)
	SaveBotChat(c context.Context, botID, botChatID string, chatEntity BotChat) error
	NewBotChatEntity(c context.Context, botID string, botChat WebhookChat, appUserID int64, botUserID string, isAccessGranted bool) BotChat
	Close(c context.Context) error // TODO: Was io.Closer, should it?
}

func NewChatID(botID, botChatID string) string {
	return botID + ":" + botChatID
}
