package bots

type BotUser interface {
	GetUserID() int64
	IsAccessGranted() bool
}

type BotUserStore interface {
	GetBotUserById(botUserId string) (BotUser, error)
	SaveBotUser(botUserId string, botUserEntity BotUser) error
	//io.Closer
}


