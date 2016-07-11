package bots

import "fmt"

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

func (t theSingleLocaleTranslator) Translate(key string, args ...interface{}) string {
	return t.Translator.Translate(key, t.locale.Code5, args...)
}

func (t theSingleLocaleTranslator) Locale() Locale {
	return t.locale
}


func (t theSingleLocaleTranslator) TranslateNoWarning(key string, args ...interface{}) string {
	return t.Translator.TranslateNoWarning(key, t.locale.Code5, args...)
}

var _ SingleLocaleTranslator = (*theSingleLocaleTranslator)(nil)

func NewSingleMapTranslator(locale Locale, translator Translator) SingleLocaleTranslator {
	return theSingleLocaleTranslator{
		locale:     locale,
		Translator: translator,
	}
}

func (t mapTranslator) _translate(warn bool, key, locale string, args ...interface{}) string {
	s, found := t.translations[key][locale]
	if !found {
		if warn {
			t.logger.Warningf("Translation not found by key & locale: key=%v&locale=%v", key, locale)
		}
		s = key
	} else if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}
	return s
}

func (t mapTranslator) Translate(key, locale string, args ...interface{}) string {
	return t._translate(true, key, locale, args...)
}

func (t mapTranslator) TranslateNoWarning(key, locale string, args ...interface{}) string {
	return t._translate(false, key, locale, args...)
}
