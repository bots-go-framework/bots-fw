package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/stretchr/testify/assert"
	"github.com/strongo/i18n"
	"os"
	"strings"
	"testing"
)

func dummyBotProfile() BotProfile {
	router := &webhooksRouter{}
	newBotChatDate := func() botsfwmodels.BotChatData {
		return nil
	}
	newBotUserData := func() botsfwmodels.PlatformUserData {
		return nil
	}
	newAppUserData := func() botsfwmodels.AppUserData {
		return nil
	}
	getAppUserByID := func(ctx context.Context, tx dal.ReadSession, botID, appUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
		return
	}
	return NewBotProfile("test", router, newBotChatDate, newBotUserData, newAppUserData, getAppUserByID, i18n.LocaleEnUS, []i18n.Locale{}, BotTranslations{})
}

func TestNewBotSettings(t *testing.T) {
	const (
		platform    = botsfwconst.PlatformTelegram
		code        = "TestBot1"
		token       = "TestToken2"
		localeCode5 = "Kode5"
		gaToken     = "ga-token1"
	)
	assertBotSettings := func(bs BotSettings) {
		assert.Equal(t, platform, bs.Platform)
		assert.Equal(t, code, bs.Code)
		assert.Equal(t, token, bs.Token)
		assert.Equal(t, localeCode5, bs.Locale.Code5)
		assert.Equal(t, gaToken, bs.GAToken)
	}

	testBotProfile := dummyBotProfile()

	getDatabase := func(_ context.Context) (db dal.DB, err error) {
		return
	}

	getAppUser := func(ctx context.Context, tx dal.ReadSession, botID, appUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
		return
	}
	t.Run("hardcoded", func(t *testing.T) {
		bs := NewBotSettings(platform, EnvLocal, testBotProfile, code, "", token, gaToken, i18n.Locale{Code5: localeCode5}, getDatabase, getAppUser)
		assertBotSettings(bs)
	})
	t.Run("from_env_vars", func(t *testing.T) {
		if err := os.Setenv("TELEGRAM_BOT_TOKEN_"+strings.ToUpper(code), token); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		if err := os.Setenv("TELEGRAM_GA_TOKEN_"+strings.ToUpper(code), gaToken); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		bs := NewBotSettings(platform, EnvLocal, testBotProfile, code, "", "", "", i18n.Locale{Code5: localeCode5}, getDatabase, getAppUser)
		assertBotSettings(bs)
	})
}

func TestNewBotSettingsBy(t *testing.T) {
	type args struct {
		bots []BotSettings
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
			name: "single_bot",
			args: args{
				bots: []BotSettings{
					{
						Profile: testBotProfile,
						Code:    "TestBot",
						ID:      "test123",
					},
				},
			},
			expectsPanic: false,
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
			actual := NewBotSettingsBy(tt.args.bots...)
			assert.Equal(t, len(tt.args.bots), len(actual.ByCode))
		})
	}
}
