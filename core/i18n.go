package bots

import (
	"fmt"
	"strings"
)

type Translator interface {
	Translate(key, locale string) string
	TranslateNoWarning(key, locale string) string
}

type SingleLocaleTranslator interface {
	Locale() Locale
	Translate(key string) string
	TranslateNoWarning(key string) string
}

type LocalesProvider interface {
	GetLocaleByCode5(code5 string) (Locale, error)
}

type Locale struct {
	Code5        string
	IsRtl        bool
	NativeTitle  string
	EnglishTitle string
	FlagIcon     string
}

func (l Locale) SiteCode() string {
	return strings.ToLower(l.Code5)
}

func (l Locale) String() string {
	return fmt.Sprintf(`Locale{Code5: "%v", IsRtl: %v, NativeTitle: "%v", EnglishTitle: "%v", FlagIcon: "%v"}`, l.Code5, l.IsRtl, l.NativeTitle, l.EnglishTitle, l.FlagIcon)
}

func (l Locale) TitleWithIcon() string {
	if l.IsRtl {
		return l.NativeTitle + " " + l.FlagIcon
	} else {
		return l.FlagIcon + " " + l.NativeTitle
	}

}

func (l Locale) TitleWithIconAndNumber(i int) string {
	if l.IsRtl {
		return fmt.Sprintf("%v %v .%d/", l.FlagIcon, l.NativeTitle, i)
	} else {
		return fmt.Sprintf("/%d. %v %v", i, l.NativeTitle, l.FlagIcon)
	}
}
