package botsfw

import "github.com/strongo/i18n"

var _ i18n.SingleLocaleTranslator = (*translator)(nil)

type translator struct {
	i18n.Translator
	localeCode5 func() string
}

func (t translator) Locale() i18n.Locale {
	return i18n.GetLocaleByCode5(t.localeCode5())
}

func (t translator) TranslateWithMap(key string, args map[string]string) string {
	locale := t.localeCode5()
	return t.Translator.TranslateWithMap(key, locale, args)
}

// Translate translates string
func (t translator) Translate(key string, args ...interface{}) string {
	locale := t.localeCode5()
	return t.Translator.Translate(key, locale, args...)
}

// TranslateNoWarning translates string without warnings
func (t translator) TranslateNoWarning(key string, args ...interface{}) string {
	locale := t.localeCode5()
	return t.Translator.TranslateNoWarning(key, locale, args...)
}
