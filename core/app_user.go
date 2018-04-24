package bots

import (
	"github.com/strongo/app"
	"context"
)

//type AppUserIntID int64

// BotAppUser holds information about bot app user
type BotAppUser interface {
	strongo.AppUser
	//GetAppUserIntID() int64
	SetBotUserID(platform, botID, botUserID string)
}

// BotAppUserStore interface for storing user information to persistent store
type BotAppUserStore interface {
	GetAppUserByID(c context.Context, appUserID int64, appUser BotAppUser) error
	CreateAppUser(c context.Context, botID string, actor WebhookActor) (appUserID int64, appUserEntity BotAppUser, err error)
	//SaveAppUser(c context.Context, appUserId int64, appUserEntity BotAppUser) error
}
