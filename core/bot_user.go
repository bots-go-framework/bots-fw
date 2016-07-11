package bots

type BotUser interface {
	GetAppUserIntID() int64
	IsAccessGranted() bool
	SetAppUserIntID(appUserID int64)
	SetDtUpdatedToNow()
}

type BotUserStore interface {
	GetBotUserById(botUserID interface{}) (BotUser, error)
	SaveBotUser(botUserID interface{}, botUserEntity BotUser) error
	CreateBotUser(apiUser WebhookActor) (BotUser, error)
	//io.Closer
}
