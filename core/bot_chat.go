package bots

import (
	"github.com/satori/go.uuid"
	"io"
)

type BotChat interface {
	GetBotID() string
	SetBotID(botID string)

	GetAppUserIntID() int64
	SetAppUserIntID(id int64)

	GetBotUserID() interface{}
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
	GetBotChatEntityById(botChatId interface{}) (BotChat, error)
	SaveBotChat(botChatId interface{}, chatEntity BotChat) error
	NewBotChatEntity(botID string, botChatId interface{}, appUserID int64, botUserID interface{}, isAccessGranted bool) BotChat
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	io.Closer
}
