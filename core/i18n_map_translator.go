package bots

type mapTranslator struct {
	translations map[string]map[string]string
	logger       Logger
}

func NewMapTranslator(translations map[string]map[string]string, logger Logger) Translator {
	return mapTranslator{translations: translations, logger: logger}
}

type theSingleLocaleTranslator struct {
	locale Locale
	Translator
}

func (t theSingleLocaleTranslator) Translate(key string) string {
	return t.Translator.Translate(key, t.locale.Code5)
}

func (t theSingleLocaleTranslator) Locale() Locale {
	return t.locale
}


func (t theSingleLocaleTranslator) TranslateNoWarning(key string) string {
	return t.Translator.TranslateNoWarning(key, t.locale.Code5)
}

var _ SingleLocaleTranslator = (*theSingleLocaleTranslator)(nil)

func NewSingleMapTranslator(locale Locale, translator Translator) SingleLocaleTranslator {
	return theSingleLocaleTranslator{
		locale:     locale,
		Translator: translator,
	}
}

func (t mapTranslator) Translate(key, locale string) string {
	value, found := t.translations[key][locale]
	if !found {
		t.logger.Warningf("Translation not found by key & locale: key=%v&locale=%v", key, locale)
		value = key
	}
	return value
}

func (t mapTranslator) TranslateNoWarning(key, locale string) string {
	value, found := t.translations[key][locale]
	if !found {
		value = key
	}
	return value
}
