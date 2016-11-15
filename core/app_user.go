package bots

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
)

//type AppUserIntID int64

type BotAppUser interface {
	strongo.AppUser
	//GetAppUserIntID() int64
	SetBotUserID(platform string, id interface{})
}

type BotAppUserStore interface {
	GetAppUserByID(c context.Context, appUserId int64, appUser BotAppUser) error
	CreateAppUser(c context.Context, actor WebhookActor) (appUserId int64, appUserEntity BotAppUser, err error)
	SaveAppUser(c context.Context, appUserId int64, appUserEntity BotAppUser) error
}
