package bots

import (
	"io"
	"github.com/strongo/app"
)

type BotChat interface {
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
	IsAwaitingReplyTo(code string, logger strongo.Logger) bool
	AddWizardParam(key, value string)
	GetWizardParam(key string) string
	PopStepsFromAwaitingReplyUpToSpecificParent(code string, logger strongo.Logger)
	PushStepToAwaitingReplyTo(code string, logger strongo.Logger)
}

type BotChatStore interface {
	GetBotChatEntityById(botChatId interface{}) (BotChat, error)
	SaveBotChat(botChatId interface{}, chatEntity BotChat) error
	NewBotChatEntity(botChatId interface{}, appUserID int64, botUserID interface{}, isAccessGranted bool) BotChat
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	io.Closer
}
