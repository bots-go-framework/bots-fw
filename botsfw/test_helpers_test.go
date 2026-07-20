package botsfw

import (
	"context"
	"fmt"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/i18n"
)

// testAppContext supplies only the presentation dependencies required by core tests.
type testAppContext struct{}

func (testAppContext) SupportedLocales() []i18n.Locale { return []i18n.Locale{i18n.LocaleEnUS} }
func (testAppContext) GetLocaleByCode5(code5 string) (i18n.Locale, error) {
	if code5 == "en-US" {
		return i18n.LocaleEnUS, nil
	}
	return i18n.Locale{}, fmt.Errorf("unsupported locale: %s", code5)
}
func (testAppContext) GetTranslator(context.Context) i18n.Translator { return testTranslator{} }
func (testAppContext) SetLocale(string) error                        { return nil }

type testTranslator struct{}

func (testTranslator) Translate(key, _ string, _ ...any) string { return key }
func (testTranslator) TranslateWithMap(key, _ string, _ map[string]string) string {
	return key
}
func (testTranslator) TranslateNoWarning(key, _ string, _ ...any) string { return key }

func newTestProfile(id string) BotProfile {
	return NewBotProfile(
		id,
		nil,
		func() botsfwmodels.BotChatData { return &botsfwmodels.ChatBaseData{} },
		func() botsfwmodels.PlatformUserData { return &botsfwmodels.PlatformUserBaseDbo{} },
		i18n.LocaleEnUS,
		nil,
		BotTranslations{},
	)
}
