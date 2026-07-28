package botswebhook

import (
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-store/botsfwstore/botsfwstoretest"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/stretchr/testify/assert"
	"github.com/strongo/i18n"
	"os"
	"strings"
	"testing"
)

func dummyBotProfile() botsfw.BotProfile {
	router := &webhooksRouter{}
	newBotChatDate := func() botsfwmodels.BotChatData {
		return nil
	}
	newBotUserData := func() botsfwmodels.PlatformUserData {
		return nil
	}
	return botsfw.NewBotProfile("test", router, newBotChatDate, newBotUserData, i18n.LocaleEnUS, []i18n.Locale{}, botsfw.BotTranslations{})
}

func TestNewBotSettings(t *testing.T) {
	const (
		platform    = botsfwconst.PlatformTelegram
		code        = "TestBot1"
		token       = "TestToken2"
		localeCode5 = "Kode5"
		gaToken     = "ga-token1"
	)
	assertBotSettings := func(bs botsfw.BotSettings) {
		assert.Equal(t, platform, bs.Platform)
		assert.Equal(t, code, bs.Code)
		assert.Equal(t, token, bs.Token)
		assert.Equal(t, localeCode5, bs.Locale.Code5)
		assert.Equal(t, gaToken, bs.GAToken) //nolint:staticcheck // intentionally testing deprecated field backward compatibility; UA was retired 2024-07-01
	}

	testBotProfile := dummyBotProfile()

	store := &botsfwstoretest.FakeStateStore{}
	t.Run("hardcoded", func(t *testing.T) {
		bs := botsfw.NewBotSettings(platform, botsfw.EnvLocal, testBotProfile, code, "", token, gaToken, i18n.Locale{Code5: localeCode5}, store)
		assertBotSettings(bs)
	})
	t.Run("from_env_vars", func(t *testing.T) {
		if err := os.Setenv("TELEGRAM_BOT_TOKEN_"+strings.ToUpper(code), token); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		if err := os.Setenv("TELEGRAM_GA_TOKEN_"+strings.ToUpper(code), gaToken); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		bs := botsfw.NewBotSettings(platform, botsfw.EnvLocal, testBotProfile, code, "", "", "", i18n.Locale{Code5: localeCode5}, store)
		assertBotSettings(bs)
	})
}

func TestNewBotSettingsBy(t *testing.T) {
	type args struct {
		bots []botsfw.BotSettings
	}

	testBotProfile := dummyBotProfile()

	tests := []struct {
		name         string
		args         args
		expectsPanic bool
	}{
		{
			name:         "no_bots",
			args:         args{},
			expectsPanic: true,
		},
		{
			// Platform is required: bot codes are unique per platform, and a bot
			// with no platform can never be resolved by the platform-scoped
			// GetBotContext, so registering one is always a mistake.
			name: "single_bot",
			args: args{
				bots: []botsfw.BotSettings{
					{
						Platform: botsfwconst.PlatformTelegram,
						Profile:  testBotProfile,
						Code:     "TestBot",
						ID:       "test123",
					},
				},
			},
			expectsPanic: false,
		},
		{
			name: "bot_without_platform",
			args: args{
				bots: []botsfw.BotSettings{
					{
						Profile: testBotProfile,
						Code:    "NoPlatformBot",
					},
				},
			},
			expectsPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectsPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewBotSettingsBy() did not panic")
					}
				}()
			}
			actual := botsfw.NewBotSettingsBy(tt.args.bots...)
			// Count via the platform-scoped map: ByCode is deprecated and
			// first-wins, so it under-counts when a code is shared across platforms.
			var registered int
			for _, byCode := range actual.ByPlatformAndCode {
				registered += len(byCode)
			}
			assert.Equal(t, len(tt.args.bots), registered)
		})
	}
}
