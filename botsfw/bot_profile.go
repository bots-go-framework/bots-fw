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
	NewBotChatData() botsfwmodels.ChatData
	NewBotUserData() botsfwmodels.BotUserData
	NewAppUserData() botsfwmodels.AppUserData
}

var _ BotProfile = (*botProfile)(nil)

type botProfile struct {
	id               string
	defaultLocale    i18n.Locale
	supportedLocales []i18n.Locale
	newBotChatData   func() botsfwmodels.ChatData
	newBotUserData   func() botsfwmodels.BotUserData
	newAppUserData   func() botsfwmodels.AppUserData
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

func (v *botProfile) NewBotChatData() botsfwmodels.ChatData {
	return v.newBotChatData()
}

func (v *botProfile) NewBotUserData() botsfwmodels.BotUserData {
	return v.newBotUserData()
}

func (v *botProfile) NewAppUserData() botsfwmodels.AppUserData {
	return v.newAppUserData()
}

func NewBotProfile(
	id string,
	router *WebhooksRouter,
	newChatData func() botsfwmodels.ChatData,
	newUserData func() botsfwmodels.BotUserData,
	newAppUserData func() botsfwmodels.AppUserData,
	defaultLocale i18n.Locale,
	supportedLocales []i18n.Locale,
) BotProfile {
	if strings.TrimSpace(id) == "" {
		panic("missing required parameter: id")
	}
	if newChatData == nil {
		panic("missing required parameter: newBotChatData")
	}
	if newUserData == nil {
		panic("missing required parameter: newBotUserData")
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
		newBotChatData:   newChatData,
		newBotUserData:   newUserData,
		newAppUserData:   newAppUserData,
	}
}
