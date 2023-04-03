package bots

import (
	"context"
	"github.com/strongo/app/user"
)

// BotUser interface provides information about bot user
type BotUser interface {
	GetAppUserIntID() int64
	IsAccessGranted() bool
	SetAccessGranted(value bool) bool
	SetAppUserIntID(appUserID int64)
	user.UpdatedTimeSetter
}

// BotUserStore provider to store information about bot user
type BotUserStore interface {
	GetBotUserByID(c context.Context, botUserID interface{}) (BotUser, error)
	SaveBotUser(c context.Context, botUserID interface{}, botUserEntity BotUser) error
	CreateBotUser(c context.Context, botID string, apiUser WebhookActor) (BotUser, error)
	//io.Closer
}
