package bots

type AppContext interface {
	AppUserEntityKind() string
	NewAppUserEntity() AppUser
}
