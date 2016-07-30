package bots

import "github.com/strongo/app"

//type AppUserIntID int64

type BotAppUser interface {
	strongo.AppUser
	//GetAppUserIntID() int64
	SetBotUserID(platform string, id interface{})
}

type BotAppUserStore interface {
	GetAppUserByID(appUserId int64, appUser BotAppUser) error
	CreateAppUser(actor WebhookActor) (appUserId int64, appUserEntity BotAppUser, err error)
	SaveAppUser(appUserId int64, appUserEntity BotAppUser) error
}
