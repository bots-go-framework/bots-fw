package bots

import (
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

type BotChat interface {
	GetBotID() string
	SetBotID(botID string)

	GetAppUserIntID() int64
	SetAppUserIntID(id int64)

	GetBotUserIntID() int
	GetBotUserStringID() string
	SetBotUserID(id interface{})

	IsAccessGranted() bool
	SetAccessGranted(value bool)

	GetPreferredLanguage() string
	SetPreferredLanguage(value string)

	SetDtUpdatedToNow()
	SetDtLastInteractionToNow()

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
	GetBotChatEntityById(c context.Context, botID, botChatID string) (BotChat, error)
	SaveBotChat(c context.Context, botID, botChatID string, chatEntity BotChat) error
	NewBotChatEntity(c context.Context, botID string, botChatId string, appUserID int64, botUserID string, isAccessGranted bool) BotChat
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	Close(c context.Context) error // TODO: Was io.Closer, should it?
}


func NewChatID(botID, botChatID string) string {
	return botID + ":" + botChatID
}