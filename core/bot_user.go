package bots

import "golang.org/x/net/context"

type BotUser interface {
	GetAppUserIntID() int64
	IsAccessGranted() bool
	SetAccessGranted(value bool)
	SetAppUserIntID(appUserID int64)
	SetDtUpdatedToNow()
}

type BotUserStore interface {
	GetBotUserById(c context.Context, botUserID interface{}) (BotUser, error)
	SaveBotUser(c context.Context, botUserID interface{}, botUserEntity BotUser) error
	CreateBotUser(c context.Context, apiUser WebhookActor) (BotUser, error)
	//io.Closer
}
