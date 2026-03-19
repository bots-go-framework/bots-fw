package botswebhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/stretchr/testify/assert"
)

func TestIsRunningLocally(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"localhost", true},
		{"abc123.ngrok.io", true},
		{"abc123.ngrok.dev", true},
		{"abc123.ngrok.app", true},
		{"abc123.ngrok-free.app", true},
		{"example.com", false},
		{"myapp.herokuapp.com", false},
		{"", false},
		{"production.api.example.com", false},
		{"localhost.example.com", false}, // not exactly "localhost"
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isRunningLocally(tt.host)
			assert.Equal(t, tt.expected, result, "isRunningLocally(%q)", tt.host)
		})
	}
}

func TestInvalidContextOrInputs(t *testing.T) {
	d := webhookDriver{}
	ctx := context.Background()

	t.Run("error_returns_true", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		result := d.invalidContextOrInputs(ctx, w, r, nil, nil, assert.AnError)
		assert.True(t, result)
	})

	t.Run("auth_failed_returns_403", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		err := botsfw.ErrAuthFailed("auth failed")
		result := d.invalidContextOrInputs(ctx, w, r, nil, nil, err)
		assert.True(t, result)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("nil_botContext_nil_inputs", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		result := d.invalidContextOrInputs(ctx, w, r, nil, nil, nil)
		assert.True(t, result)
	})

	t.Run("nil_botContext_empty_inputs", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		result := d.invalidContextOrInputs(ctx, w, r, nil, []botinput.EntryInputs{}, nil)
		assert.True(t, result)
	})

	t.Run("nil_botContext_with_inputs", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		entries := []botinput.EntryInputs{{}}
		result := d.invalidContextOrInputs(ctx, w, r, nil, entries, nil)
		assert.True(t, result)
	})

	t.Run("valid_botContext_nil_inputs", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		bc := &botsfw.BotContext{
			BotSettings: &botsfw.BotSettings{Env: "production"},
		}
		result := d.invalidContextOrInputs(ctx, w, r, bc, nil, nil)
		assert.True(t, result)
	})

	t.Run("local_env_with_localhost", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
		bc := &botsfw.BotContext{
			BotSettings: &botsfw.BotSettings{Env: botsfw.EnvLocal},
		}
		entries := []botinput.EntryInputs{{}}
		result := d.invalidContextOrInputs(ctx, w, r, bc, entries, nil)
		assert.False(t, result)
	})

	t.Run("local_env_with_production_host_rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
		bc := &botsfw.BotContext{
			BotSettings: &botsfw.BotSettings{Env: botsfw.EnvLocal},
		}
		entries := []botinput.EntryInputs{{}}
		result := d.invalidContextOrInputs(ctx, w, r, bc, entries, nil)
		assert.True(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("production_env_with_localhost_rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "http://localhost/webhook", nil)
		bc := &botsfw.BotContext{
			BotSettings: &botsfw.BotSettings{Env: botsfw.EnvProduction},
		}
		entries := []botinput.EntryInputs{{}}
		result := d.invalidContextOrInputs(ctx, w, r, bc, entries, nil)
		assert.True(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("production_env_with_production_host", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "http://myapp.example.com/webhook", nil)
		bc := &botsfw.BotContext{
			BotSettings: &botsfw.BotSettings{Env: botsfw.EnvProduction},
		}
		entries := []botinput.EntryInputs{{}}
		result := d.invalidContextOrInputs(ctx, w, r, bc, entries, nil)
		assert.False(t, result)
	})
}

func TestNewWebhookDriver(t *testing.T) {
	t.Run("panics_with_nil_botHost", func(t *testing.T) {
		assert.PanicsWithValue(t, "required argument botHost == nil", func() {
			NewWebhookDriver(AnalyticsSettings{}, nil, "")
		})
	})

	t.Run("valid_args", func(t *testing.T) {
		host := testBotHostForDriver{}
		d := NewWebhookDriver(AnalyticsSettings{GaTrackingID: "UA-123"}, host, "footer text")
		assert.NotNil(t, d)
	})
}

type testBotHostForDriver struct{}

func (testBotHostForDriver) Context(r *http.Request) context.Context { return r.Context() }
func (testBotHostForDriver) GetHTTPClient(_ context.Context) *http.Client {
	return http.DefaultClient
}
