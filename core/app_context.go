package bots

import "reflect"

type AppContext interface {
	AppUserEntityKind() string
	AppUserEntityType() reflect.Type
	NewAppUserEntity() AppUser
	GetTranslator(l Logger) Translator
	SupportedLocales() LocalesProvider
}
