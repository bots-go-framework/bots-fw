package botsfw

import (
	"github.com/stretchr/testify/assert"
	strongo "github.com/strongo/app"
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
		bs := NewBotSettings(platform, strongo.EnvLocal, "unit-test", code, "", token, gaToken, strongo.Locale{Code5: localeCode5})
		assertBotSettings(bs)
	})
	t.Run("from_env_vars", func(t *testing.T) {
		if err := os.Setenv("TELEGRAM_BOT_TOKEN_"+strings.ToUpper(code), token); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		if err := os.Setenv("TELEGRAM_GA_TOKEN_"+strings.ToUpper(code), gaToken); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		bs := NewBotSettings(platform, strongo.EnvLocal, "unit-test", code, "", "", "", strongo.Locale{Code5: localeCode5})
		assertBotSettings(bs)
	})
}
