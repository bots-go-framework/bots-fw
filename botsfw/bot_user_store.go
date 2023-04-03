package botsfw

import "context"

//type UserID interface {
//	int | string
//}

// BotUserStore provider to store information about bot user
type BotUserStore interface {

	// GetBotUserByID returns bot user data
	GetBotUserByID(c context.Context, botUserID any) (BotUser, error)

	// SaveBotUser saves bot user data
	SaveBotUser(c context.Context, botUserID any, botUserData BotUser) error

	// CreateBotUser creates new bot user in DB
	// Deprecated: should be moved to bots-fw-* package
	CreateBotUser(c context.Context, botID string, apiUser WebhookActor) (BotUser, error)
	//io.Closer
}
