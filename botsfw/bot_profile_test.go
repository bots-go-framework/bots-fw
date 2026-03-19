package botsfw

import (
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strongo/i18n"
)

func TestBotCommand_Validate(t *testing.T) {
	t.Run("empty_command", func(t *testing.T) {
		cmd := BotCommand{Command: "", Description: "desc"}
		err := cmd.Validate()
		require.Error(t, err)
		assert.Equal(t, "command is required", err.Error())
	})

	t.Run("command_too_long", func(t *testing.T) {
		cmd := BotCommand{Command: "abcdefghijklmnopqrstuvwxyz1234567", Description: "desc"}
		assert.True(t, len(cmd.Command) > 32)
		err := cmd.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command is too long")
	})

	t.Run("empty_description", func(t *testing.T) {
		cmd := BotCommand{Command: "help", Description: ""}
		err := cmd.Validate()
		require.Error(t, err)
		assert.Equal(t, "description is required", err.Error())
	})

	t.Run("description_too_long", func(t *testing.T) {
		longDesc := make([]byte, 257)
		for i := range longDesc {
			longDesc[i] = 'a'
		}
		cmd := BotCommand{Command: "help", Description: string(longDesc)}
		err := cmd.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description is too long")
	})

	t.Run("valid", func(t *testing.T) {
		cmd := BotCommand{Command: "help", Description: "Show help info"}
		err := cmd.Validate()
		assert.NoError(t, err)
	})

	t.Run("max_length_command_valid", func(t *testing.T) {
		cmd := BotCommand{Command: "abcdefghijklmnopqrstuvwxyz123456", Description: "ok"}
		assert.Equal(t, 32, len(cmd.Command))
		assert.NoError(t, cmd.Validate())
	})

	t.Run("max_length_description_valid", func(t *testing.T) {
		desc := make([]byte, 256)
		for i := range desc {
			desc[i] = 'x'
		}
		cmd := BotCommand{Command: "test", Description: string(desc)}
		assert.NoError(t, cmd.Validate())
	})
}

func TestNewBotProfile(t *testing.T) {
	newBotChatData := func() botsfwmodels.BotChatData { return nil }
	newBotUserData := func() botsfwmodels.PlatformUserData { return nil }
	newAppUserData := func() botsfwmodels.AppUserData { return nil }
	locale := i18n.LocaleEnUS

	t.Run("panics_on_empty_id", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotProfile("", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, nil, BotTranslations{})
		})
	})

	t.Run("panics_on_whitespace_id", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotProfile("   ", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, nil, BotTranslations{})
		})
	})

	t.Run("panics_on_nil_newBotChatData", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotProfile("testbot", nil, nil, newBotUserData, newAppUserData, nil, locale, nil, BotTranslations{})
		})
	})

	t.Run("panics_on_nil_newBotUserData", func(t *testing.T) {
		assert.Panics(t, func() {
			NewBotProfile("testbot", nil, newBotChatData, nil, newAppUserData, nil, locale, nil, BotTranslations{})
		})
	})

	t.Run("valid_construction", func(t *testing.T) {
		translations := BotTranslations{
			Description: "Test bot",
			Commands:    []BotCommand{{Command: "start", Description: "Start the bot"}},
		}
		profile := NewBotProfile("mybot", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, nil, translations)
		require.NotNil(t, profile)
		assert.Equal(t, "mybot", profile.ID())
		assert.Nil(t, profile.Router())
		assert.Equal(t, locale, profile.DefaultLocale())
		assert.Equal(t, translations, profile.GetTranslations())
	})

	t.Run("default_locale_added_to_supported", func(t *testing.T) {
		profile := NewBotProfile("bot1", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, nil, BotTranslations{})
		supported := profile.SupportedLocales()
		require.Len(t, supported, 1)
		assert.Equal(t, locale.Code5, supported[0].Code5)
	})

	t.Run("default_locale_not_duplicated", func(t *testing.T) {
		profile := NewBotProfile("bot2", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, []i18n.Locale{locale}, BotTranslations{})
		supported := profile.SupportedLocales()
		assert.Len(t, supported, 1)
	})

	t.Run("multiple_supported_locales", func(t *testing.T) {
		otherLocale := i18n.Locale{Code5: "de-DE"}
		profile := NewBotProfile("bot3", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, []i18n.Locale{otherLocale}, BotTranslations{})
		supported := profile.SupportedLocales()
		assert.Len(t, supported, 2) // otherLocale + defaultLocale appended
	})

	t.Run("accessor_methods", func(t *testing.T) {
		profile := NewBotProfile("bot4", nil, newBotChatData, newBotUserData, newAppUserData, nil, locale, nil, BotTranslations{})
		// NewBotChatData
		assert.Nil(t, profile.NewBotChatData())
		// NewPlatformUserData
		assert.Nil(t, profile.NewPlatformUserData())
		// NewAppUserData
		assert.Nil(t, profile.NewAppUserData())
	})
}
