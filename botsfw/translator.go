package botsfw

import "github.com/strongo/i18n"

type translator struct {
	i18n.Translator
	localeCode5 func() string
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
