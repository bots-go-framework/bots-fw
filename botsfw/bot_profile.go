package botsfw

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/i18n"
	"strings"
)

type BotProfile interface {
	ID() string
	Router() *WebhooksRouter
	DefaultLocale() i18n.Locale
	SupportedLocales() []i18n.Locale
	NewChatData() botsfwmodels.ChatData
	NewUserData() botsfwmodels.AppUserData
}

var _ BotProfile = (*botProfile)(nil)

type botProfile struct {
	id               string
	defaultLocale    i18n.Locale
	supportedLocales []i18n.Locale
	newChatData      func() botsfwmodels.ChatData
	newUserData      func() botsfwmodels.AppUserData
	router           *WebhooksRouter
}

func (v *botProfile) ID() string {
	return v.id
}

func (v *botProfile) Router() *WebhooksRouter {
	return v.router
}

func (v *botProfile) DefaultLocale() i18n.Locale {
	return v.defaultLocale
}

func (v *botProfile) SupportedLocales() []i18n.Locale {
	return v.supportedLocales[:]
}

func (v *botProfile) NewChatData() botsfwmodels.ChatData {
	return v.newChatData()
}

func (v *botProfile) NewUserData() botsfwmodels.AppUserData {
	return v.newUserData()
}

func NewBotProfile(
	id string,
	router *WebhooksRouter,
	newChatData func() botsfwmodels.ChatData,
	newUserData func() botsfwmodels.AppUserData,
	defaultLocale i18n.Locale,
	supportedLocales []i18n.Locale,
) BotProfile {
	if strings.TrimSpace(id) == "" {
		panic("missing required parameter: id")
	}
	if newChatData == nil {
		panic("missing required parameter: newChatData")
	}
	if newUserData == nil {
		panic("missing required parameter: newUserData")
	}
	var defaultLocaleInSupportedLocales bool
	for _, locale := range supportedLocales {
		if locale.Code5 == defaultLocale.Code5 {
			defaultLocaleInSupportedLocales = true
			break
		}
	}
	if !defaultLocaleInSupportedLocales {
		supportedLocales = append(supportedLocales, defaultLocale)
	}
	return &botProfile{
		id:               id,
		router:           router,
		defaultLocale:    defaultLocale,
		supportedLocales: supportedLocales,
		newChatData:      newChatData,
		newUserData:      newUserData,
	}
}
