package botsfw

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/i18n"
)

// testAppContext is a minimal implementation of AppContext for testing.
type testAppContext struct{}

func (testAppContext) SupportedLocales() []i18n.Locale {
	return []i18n.Locale{i18n.LocaleEnUS}
}

func (testAppContext) GetLocaleByCode5(code5 string) (i18n.Locale, error) {
	if code5 == "en-US" {
		return i18n.LocaleEnUS, nil
	}
	return i18n.Locale{}, fmt.Errorf("unsupported locale: %s", code5)
}

func (testAppContext) GetTranslator(_ context.Context) i18n.Translator {
	return testTranslator{}
}

func (testAppContext) SetLocale(_ string) error {
	return nil
}

func (testAppContext) CreateAppUserFromBotUser(_ context.Context, _ dal.ReadwriteTransaction, _ botsdal.Bot) (
	record.DataWithID[string, botsfwmodels.AppUserData], botsdal.BotUser, error,
) {
	panic("not implemented in test")
}

// testTranslator is a simple translator that returns the key as the translation.
type testTranslator struct{}

func (testTranslator) Translate(key, _ string, args ...any) string {
	return key
}
func (testTranslator) TranslateWithMap(key, _ string, _ map[string]string) string {
	return key
}
func (testTranslator) TranslateNoWarning(key, _ string, args ...any) string {
	return key
}
