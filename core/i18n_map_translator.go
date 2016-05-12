package bots

type mapTranslator struct {
	translations map[string]map[string]string
	logger       Logger
}

func NewMapTranslator(translations map[string]map[string]string, logger Logger) Translator {
	return mapTranslator{translations: translations, logger: logger}
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
