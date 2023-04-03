package bots

import (
	"context"
	"github.com/strongo/app"
)

//type AppUserIntID int64

// BotAppUser holds information about bot app user
type BotAppUser interface {
	strongo.AppUser
	SetBotUserID(platform, botID, botUserID string)
	GetFullName() string
}

// BotAppUserStore interface for storing user information to persistent store
type BotAppUserStore interface {
	GetAppUserByID(c context.Context, appUserID int64, appUser BotAppUser) error
	CreateAppUser(c context.Context, botID string, actor WebhookActor) (appUserID int64, appUserEntity BotAppUser, err error)
	//SaveAppUser(c context.Context, appUserId int64, appUserEntity BotAppUser) error
}
