package bots

type AppContext interface {
	AppUserEntityKind() string
	NewAppUserEntity() AppUser
	GetTranslator(l Logger) Translator
	SupportedLocales() LocalesProvider
}
