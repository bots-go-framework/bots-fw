package bots

//type AppUserIntID int64

type AppUser interface {
	//GetAppUserIntID() int64
	SetPreferredLocale(code5 string) error
	PreferredLocale() string

	SetNames(firs, last, user string)
	SetBotUserID(platform string, id interface{})
}

type AppUserStore interface {
	GetAppUserByID(appUserId int64, appUser AppUser) error
	CreateAppUser(actor WebhookActor) (appUserId int64, appUserEntity AppUser, err error)
	SaveAppUser(appUserId int64, appUserEntity AppUser) error
}
