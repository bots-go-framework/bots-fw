package botsfw

import (
	"github.com/stretchr/testify/assert"
	strongo "github.com/strongo/app"
	"github.com/strongo/i18n"
	"os"
	"strings"
	"testing"
)

func TestNewBotSettings(t *testing.T) {
	const (
		platform    = PlatformTelegram
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
	t.Run("hardcoded", func(t *testing.T) {
		bs := NewBotSettings(platform, strongo.EnvLocal, "unit-test", code, "", token, gaToken, i18n.Locale{Code5: localeCode5})
		assertBotSettings(bs)
	})
	t.Run("from_env_vars", func(t *testing.T) {
		if err := os.Setenv("TELEGRAM_BOT_TOKEN_"+strings.ToUpper(code), token); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		if err := os.Setenv("TELEGRAM_GA_TOKEN_"+strings.ToUpper(code), gaToken); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		bs := NewBotSettings(platform, strongo.EnvLocal, "unit-test", code, "", "", "", i18n.Locale{Code5: localeCode5})
		assertBotSettings(bs)
	})
}

func TestNewBotSettingsBy(t *testing.T) {
	type args struct {
		getRouter func(profile string) WebhooksRouter
		bots      []BotSettings
	}
	tests := []struct {
		name         string
		args         args
		expectsPanic bool
	}{
		{
			name: "no_bots",
			args: args{
				getRouter: func(profile string) WebhooksRouter {
					return WebhooksRouter{}
				},
			},
			expectsPanic: true,
		},
		{
			name: "single_bot",
			args: args{
				getRouter: func(profile string) WebhooksRouter {
					return WebhooksRouter{}
				},
				bots: []BotSettings{
					{
						Profile: "test",
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
			actual := NewBotSettingsBy(tt.args.getRouter, tt.args.bots...)
			assert.Equal(t, len(tt.args.bots), len(actual.ByCode))
		})
	}
}
