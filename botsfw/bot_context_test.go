package botsfw

import (
	"context"
	"net/http"
	"testing"

	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/stretchr/testify/assert"
)

type testBotHost struct{}

func (testBotHost) Context(r *http.Request) context.Context { return r.Context() }
func (testBotHost) GetHTTPClient(_ context.Context) *http.Client {
	return http.DefaultClient
}

func TestNewBotContext(t *testing.T) {
	t.Run("panics_on_nil_botHost", func(t *testing.T) {
		assert.PanicsWithValue(t, "required argument botHost is nil", func() {
			NewBotContext(nil, &BotSettings{Code: "test"})
		})
	})

	t.Run("panics_on_nil_botSettings", func(t *testing.T) {
		assert.PanicsWithValue(t, "required argument botSettings is nil", func() {
			NewBotContext(testBotHost{}, nil)
		})
	})

	t.Run("panics_on_empty_code", func(t *testing.T) {
		assert.PanicsWithValue(t, "ReferredTo botSettings.Code is empty string", func() {
			NewBotContext(testBotHost{}, &BotSettings{})
		})
	})

	t.Run("valid_args", func(t *testing.T) {
		settings := &BotSettings{Code: "mybot"}
		bc := NewBotContext(testBotHost{}, settings)
		assert.NotNil(t, bc)
		assert.Equal(t, settings, bc.BotSettings)
		assert.NotNil(t, bc.BotHost)
	})
}

type testBotSettingsProvider struct {
	settingsBy BotSettingsBy
}

func (p testBotSettingsProvider) provide(_ context.Context) BotSettingsBy {
	return p.settingsBy
}

func TestNewBotContextProvider(t *testing.T) {
	host := testBotHost{}
	// Platform is required: GetBotContext is platform-scoped, and NewBotSettings
	// panics without one. A BotSettings with a blank Platform is malformed and no
	// longer resolves for any platform.
	settings := &BotSettings{Code: "mybot", Platform: botsfwconst.PlatformTelegram}
	settingsBy := BotSettingsBy{
		ByCode: map[string]*BotSettings{"mybot": settings},
		ByID:   map[string]*BotSettings{},
	}
	provider := testBotSettingsProvider{settingsBy: settingsBy}

	t.Run("panics_on_nil_botHost", func(t *testing.T) {
		assert.PanicsWithValue(t, "required argument botHost == nil", func() {
			NewBotContextProvider(nil, nil, nil)
		})
	})

	t.Run("panics_on_nil_appContext", func(t *testing.T) {
		assert.PanicsWithValue(t, "required argument appContext == nil", func() {
			NewBotContextProvider(host, nil, nil)
		})
	})

	t.Run("panics_on_nil_settingsProvider", func(t *testing.T) {
		mockAppCtx := &testAppContext{}
		assert.PanicsWithValue(t, "required argument botSettingProvider == nil", func() {
			NewBotContextProvider(host, mockAppCtx, nil)
		})
	})

	t.Run("valid_construction_and_lookup", func(t *testing.T) {
		mockAppCtx := &testAppContext{}
		bcp := NewBotContextProvider(host, mockAppCtx, provider.provide)
		assert.NotNil(t, bcp)

		ctx := context.Background()
		bc, err := bcp.GetBotContext(ctx, "telegram", "mybot")
		assert.NoError(t, err)
		assert.NotNil(t, bc)
		assert.Equal(t, "mybot", bc.BotSettings.Code)
	})

	t.Run("unknown_bot_returns_error", func(t *testing.T) {
		mockAppCtx := &testAppContext{}
		bcp := NewBotContextProvider(host, mockAppCtx, provider.provide)
		ctx := context.Background()
		_, err := bcp.GetBotContext(ctx, "telegram", "nonexistent")
		assert.ErrorIs(t, err, ErrUnknownBot)
	})
}

// TestGetBotContext_isPlatformScoped is the security regression test.
//
// Before this was scoped, GetBotContext accepted platformID and ignored it,
// resolving purely by ID or code. With Telegram as the only platform that was
// invisible. With two, a WhatsApp webhook could resolve a Telegram bot's settings
// — including its Token and WebhookSecretToken.
func TestGetBotContext_isPlatformScoped(t *testing.T) {
	tgBot := BotSettings{
		Platform: botsfwconst.PlatformTelegram, Code: "debtus", ID: "debtus",
		Token: "TELEGRAM-SECRET-TOKEN", Profile: newTestProfile("debtus-profile"),
	}
	waBot := BotSettings{
		Platform: botsfwconst.PlatformWhatsApp, Code: "debtus", ID: "debtus",
		Token: "WHATSAPP-SECRET-TOKEN", Profile: newTestProfile("debtus-profile"),
	}

	// The same product on two platforms shares a code. This must be registrable —
	// it used to panic with "Bot with duplicate code".
	settingsBy := NewBotSettingsBy(tgBot, waBot)
	provider := testBotSettingsProvider{settingsBy: settingsBy}
	bcp := NewBotContextProvider(testBotHost{}, &testAppContext{}, provider.provide)
	ctx := context.Background()

	t.Run("each platform resolves its own bot", func(t *testing.T) {
		tg, err := bcp.GetBotContext(ctx, botsfwconst.PlatformTelegram, "debtus")
		assert.NoError(t, err)
		assert.Equal(t, "TELEGRAM-SECRET-TOKEN", tg.BotSettings.Token)

		wa, err := bcp.GetBotContext(ctx, botsfwconst.PlatformWhatsApp, "debtus")
		assert.NoError(t, err)
		assert.Equal(t, "WHATSAPP-SECRET-TOKEN", wa.BotSettings.Token,
			"a WhatsApp webhook must never be handed Telegram's token")
	})

	t.Run("a bot on another platform is unknown, not silently substituted", func(t *testing.T) {
		_, err := bcp.GetBotContext(ctx, botsfwconst.Platform("viber"), "debtus")
		assert.ErrorIs(t, err, ErrUnknownBot)
	})

	t.Run("blank platform is rejected", func(t *testing.T) {
		_, err := bcp.GetBotContext(ctx, "", "debtus")
		assert.Error(t, err)
	})
}

// TestGetBotContext_legacyFlatMapsStillCheckPlatform pins that the compatibility
// fallback for a hand-built BotSettingsBy cannot reintroduce the leak it exists to
// be compatible with.
func TestGetBotContext_legacyFlatMapsStillCheckPlatform(t *testing.T) {
	tgOnly := &BotSettings{Platform: botsfwconst.PlatformTelegram, Code: "mybot"}
	settingsBy := BotSettingsBy{ // built by hand, no platform maps
		ByCode: map[string]*BotSettings{"mybot": tgOnly},
		ByID:   map[string]*BotSettings{},
	}
	bcp := NewBotContextProvider(testBotHost{}, &testAppContext{},
		testBotSettingsProvider{settingsBy: settingsBy}.provide)
	ctx := context.Background()

	bc, err := bcp.GetBotContext(ctx, botsfwconst.PlatformTelegram, "mybot")
	assert.NoError(t, err, "the legacy fallback must still resolve its own platform")
	assert.Equal(t, "mybot", bc.BotSettings.Code)

	_, err = bcp.GetBotContext(ctx, botsfwconst.PlatformWhatsApp, "mybot")
	assert.ErrorIs(t, err, ErrUnknownBot,
		"the legacy fallback must NOT hand a Telegram bot to a WhatsApp webhook")
}

// TestNewBotSettingsBy_duplicateCodeAcrossPlatforms pins that codes are unique per
// platform rather than globally — the change that makes a second platform
// registrable at all.
func TestNewBotSettingsBy_duplicateCodeAcrossPlatforms(t *testing.T) {
	tg := BotSettings{Platform: botsfwconst.PlatformTelegram, Code: "debtus", Profile: newTestProfile("debtus-profile")}
	wa := BotSettings{Platform: botsfwconst.PlatformWhatsApp, Code: "debtus", Profile: newTestProfile("debtus-profile")}

	assert.NotPanics(t, func() { NewBotSettingsBy(tg, wa) },
		"the same code on two platforms is legitimate")

	assert.Panics(t, func() { NewBotSettingsBy(tg, tg) },
		"a duplicate code WITHIN a platform is still a bug")

	assert.Panics(t, func() {
		NewBotSettingsBy(BotSettings{Code: "nope", Profile: newTestProfile("debtus-profile")})
	}, "a bot with no platform is malformed")
}
