package bots

type BotUser interface {
	GetAppUserID() int64
	IsAccessGranted() bool
	SetAppUserID(appUserID int64)
}

type BotUserStore interface {
	GetBotUserById(botUserID interface{}) (BotUser, error)
	SaveBotUser(botUserID interface{}, botUserEntity BotUser) error
	CreateBotUser(apiUser WebhookActor) (BotUser, error)
	//io.Closer
}


