package bots

type AppUser interface {
	//GetAppUserID() int64
	SetPreferredLocale(code5 string) error
	PreferredLocale() string
}

type AppUserStore interface {
	GetAppUserByID(appUserId int64, appUser AppUser) error
	CreateAppUser(appUserEntity AppUser) (appUserId int64, err error)
	SaveAppUser(appUserId int64, appUserEntity AppUser) error
}