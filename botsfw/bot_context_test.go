package botsfw

import (
	"context"
	"net/http"
	"testing"

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
	settings := &BotSettings{Code: "mybot"}
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
