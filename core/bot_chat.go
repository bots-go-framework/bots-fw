package bots

import "io"

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

	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(path string)
	IsAwaitingReplyTo(code string, logger Logger) bool
	AddWizardParam(name, value string, logger Logger)
	PopStepsFromAwaitingReplyToUpTo(code string, logger Logger)
	PushStepToAwaitingReplyTo(code string, logger Logger)
}

type BotChatStore interface {
	GetBotChatEntityById(botChatId interface{}) (BotChat, error)
	SaveBotChat(botChatId interface{}, chatEntity BotChat) error
	NewBotChatEntity(botChatId interface{}, appUserID int64, botUserID interface{}, isAccessGranted bool) BotChat
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	io.Closer
}
