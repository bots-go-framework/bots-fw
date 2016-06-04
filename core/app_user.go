package bots

//type AppUserID int64

type AppUser interface {
	//GetAppUserID() int64
	SetPreferredLocale(code5 string) error
	PreferredLocale() string

	SetNames(firs, last, user string)
}

type AppUserStore interface {
	GetAppUserByID(appUserId int64, appUser AppUser) error
	CreateAppUser(actor WebhookActor) (appUserId int64, appUserEntity AppUser, err error)
	SaveAppUser(appUserId int64, appUserEntity AppUser) error
}
