package bots

import "io"

type BotChat interface {
	GetAppUserID() int64
	SetAppUserID(id int64)

	GetBotUserID() interface{}
	SetBotUserID(id interface{})

	IsAccessGranted() bool
	SetAccessGranted(value bool)

	GetPreferredLanguage() string
	SetPreferredLanguage(value string)

	SetDtUpdatedToNow()

	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(string)
	IsAwaitingReplyTo(code string, logger Logger) bool
	AddWizardParam(name, value string)
	AddStepToAwaitingReplyTo(code string)
}

type BotChatStore interface {
	GetBotChatEntityById(botChatId interface{}) (BotChat, error)
	SaveBotChat(botChatId interface{}, chatEntity BotChat) error
	NewBotChatEntity(botChatId interface{}, appUserID int64, botUserID interface{}, isAccessGranted bool) BotChat
	//AddChat(chat BotChat)
	//RemoveChat(chat BotChat)
	io.Closer
}
