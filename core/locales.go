package bots

const (
	LOCALE_EN_US = "en-US"
	LOCALE_RU_RU = "ru-RU"
	LOCALE_ID_ID = "id-ID"
	LOCALE_FA_IR = "fa-IR"
	LOCALE_IT_IT = "it-IT"

	LOCALE_DE_DE = "de-DE"
	LOCALE_ES_ES = "es-ES"
	LOCALE_FR_FR = "fr-FR"
	LOCALE_PT_PT = "pt-PT"
	LOCALE_PT_BR = "pt-BR"
)

//"4. French ",
//"5. Spanish ",
//"6. Italian \xF0\x9F\x87\xAE\xF0\x9F\x87\xB9",

var LocaleEnUs = Locale{Code5: LOCALE_EN_US, NativeTitle: "English", EnglishTitle: "English", FlagIcon: "🇺🇸"}
var LocaleRuRu = Locale{Code5: LOCALE_RU_RU, NativeTitle: "Русский", EnglishTitle: "Russian", FlagIcon: "🇷🇺"}
var LocaleIdId = Locale{Code5: LOCALE_ID_ID, NativeTitle: "Indonesian", EnglishTitle: "Indonesian", FlagIcon: "🇮🇩"}
var LocaleDeDe = Locale{Code5: LOCALE_DE_DE, NativeTitle: "Deutsche", EnglishTitle: "German", FlagIcon: "🇩🇪"}
var LocaleEsEs = Locale{Code5: LOCALE_ES_ES, NativeTitle: "Español", EnglishTitle: "Spanish", FlagIcon: "🇪🇸"}
var LocaleFrFr = Locale{Code5: LOCALE_FR_FR, NativeTitle: "Français", EnglishTitle: "France", FlagIcon: "🇫🇷"}
var LocaleItIt = Locale{Code5: LOCALE_IT_IT, NativeTitle: "Italiano", EnglishTitle: "Italian", FlagIcon: "🇮🇹"}
var LocalePtPt = Locale{Code5: LOCALE_PT_PT, NativeTitle: "Português (PT)", EnglishTitle: "Portuguese (PT)", FlagIcon: "🇵🇹"}
var LocalePtBr = Locale{Code5: LOCALE_PT_BR, NativeTitle: "Português (BR)", EnglishTitle: "Portuguese (BR)", FlagIcon: "🇧🇷"}
var LocaleFaIr = Locale{Code5: LOCALE_FA_IR, IsRtl: true, NativeTitle: "فارسی", EnglishTitle: "Persian", FlagIcon: "🇮🇷"}
