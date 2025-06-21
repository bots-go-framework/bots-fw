package botsfw

import (
	"errors"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/i18n"
	"strings"
)

type BotTranslations struct {
	Description      string
	ShortDescription string
	Commands         []BotCommand
}

type BotCommand struct {
	Command     string `json:"command"`     // Text of the command; 1-32 characters. Can contain only lowercase English letters, digits and underscores.
	Description string `json:"description"` // Description of the command; 1-256 characters.
}

func (v BotCommand) Validate() error {
	if len(v.Command) == 0 {
		return errors.New("command is required")
	}
	if len(v.Command) > 32 {
		return errors.New("command is too long, expected to be 32 characters max")
	}
	if len(v.Description) == 0 {
		return errors.New("description is required")
	}
	if len(v.Description) > 256 {
		return errors.New("description is too long, expected to be 256 characters max")
	}
	return nil
}

type BotProfile interface {
	ID() string
	Router() Router
	DefaultLocale() i18n.Locale
	SupportedLocales() []i18n.Locale
	NewBotChatData() botsfwmodels.BotChatData
	NewPlatformUserData() botsfwmodels.PlatformUserData
	NewAppUserData() botsfwmodels.AppUserData // TODO: Can we get rit of it and instead use GetAppUserByID/CreateAppUser?
	GetTranslations() BotTranslations
}

var _ BotProfile = (*botProfile)(nil)

type botProfile struct {
	id               string
	defaultLocale    i18n.Locale
	supportedLocales []i18n.Locale
	newBotChatData   func() botsfwmodels.BotChatData
	newBotUserData   func() botsfwmodels.PlatformUserData
	newAppUserData   func() botsfwmodels.AppUserData
	getAppUserByID   AppUserGetter
	router           Router
	translations     BotTranslations
}

func (v *botProfile) ID() string {
	return v.id
}

func (v *botProfile) Router() Router {
	return v.router
}

func (v *botProfile) DefaultLocale() i18n.Locale {
	return v.defaultLocale
}

func (v *botProfile) SupportedLocales() []i18n.Locale {
	return v.supportedLocales[:]
}

func (v *botProfile) NewBotChatData() botsfwmodels.BotChatData {
	return v.newBotChatData()
}

func (v *botProfile) NewPlatformUserData() botsfwmodels.PlatformUserData {
	return v.newBotUserData()
}

func (v *botProfile) NewAppUserData() botsfwmodels.AppUserData {
	return v.newAppUserData()
}

func (v *botProfile) GetTranslations() BotTranslations {
	return v.translations
}

func NewBotProfile(
	id string,
	router Router,
	newBotChatData func() botsfwmodels.BotChatData,
	newBotUserData func() botsfwmodels.PlatformUserData,
	newAppUserData func() botsfwmodels.AppUserData,
	getAppUserByID AppUserGetter,
	defaultLocale i18n.Locale,
	supportedLocales []i18n.Locale,
	translations BotTranslations,
) BotProfile {
	if strings.TrimSpace(id) == "" {
		panic("missing required parameter: id")
	}
	if newBotChatData == nil {
		panic("missing required parameter: newBotChatData")
	}
	if newBotUserData == nil {
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
		newBotChatData:   newBotChatData,
		newBotUserData:   newBotUserData,
		newAppUserData:   newAppUserData,
		getAppUserByID:   getAppUserByID,
		translations:     translations,
	}
}
