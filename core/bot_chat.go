package bots

import "io"

type BotChat interface {
	GetAppUserID() int64
	SetAppUserID(id int64)

	GetBotUserID() int64
	SetBotUserID(id int64)

	IsAccessGranted() bool
	SetAccessGranted(value bool)

	GetPreferredLanguage() string
	SetPreferredLanguage(value string)

	SetDtUpdatedToNow()

	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(string)
	IsAwaitingReplyTo(code string) bool
	AddWizardParam(name, value string)
	AddStepToAwaitingReplyTo(code string)
}


type BotChatStore interface {
	GetBotChatEntityById(botChatId interface{}) (BotChat, error)
	SaveBotChat(botChatId interface{}, chatEntity BotChat) error
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	io.Closer
}

