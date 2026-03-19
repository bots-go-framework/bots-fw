package botsfw_test

import (
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"go.uber.org/mock/gomock"
)

func TestSubInterfaceMocksCanBeInstantiated(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("MockWebhookRequestContext", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookRequestContext(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookRequestContext returned nil")
		}
	})

	t.Run("MockWebhookInputContext", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookInputContext(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookInputContext returned nil")
		}
	})

	t.Run("MockWebhookUserData", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookUserData(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookUserData returned nil")
		}
	})

	t.Run("MockWebhookI18n", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookI18n(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookI18n returned nil")
		}
	})

	t.Run("MockWebhookMessaging", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookMessaging(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookMessaging returned nil")
		}
	})

	t.Run("MockWebhookTelemetry", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookTelemetry(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookTelemetry returned nil")
		}
	})

	t.Run("MockWebhookContext", func(t *testing.T) {
		m := mock_botsfw.NewMockWebhookContext(ctrl)
		if m == nil {
			t.Fatal("NewMockWebhookContext returned nil")
		}
	})
}

// TestDecompositionPolymorphism verifies that a function accepting a narrow sub-interface
// can receive a full WebhookContext mock (proving the decomposition is correct).
func TestDecompositionPolymorphism(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("WebhookContext_satisfies_WebhookTelemetry", func(t *testing.T) {
		consumeTelemetry := func(wt botsfw.WebhookTelemetry) botsfw.WebhookAnalytics {
			return wt.Analytics()
		}
		mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
		mockAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
		mockWHC.EXPECT().Analytics().Return(mockAnalytics)
		result := consumeTelemetry(mockWHC)
		if result != mockAnalytics {
			t.Error("Expected mock analytics to be returned")
		}
	})

	t.Run("WebhookContext_satisfies_WebhookRequestContext", func(t *testing.T) {
		consumeRequestCtx := func(rc botsfw.WebhookRequestContext) string {
			return rc.Environment()
		}
		mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
		mockWHC.EXPECT().Environment().Return("test-env")
		result := consumeRequestCtx(mockWHC)
		if result != "test-env" {
			t.Errorf("Expected 'test-env', got %q", result)
		}
	})

	t.Run("WebhookContext_satisfies_WebhookMessaging", func(t *testing.T) {
		consumeMessaging := func(wm botsfw.WebhookMessaging) string {
			m := wm.NewMessage("hello")
			return m.Text
		}
		mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
		mockWHC.EXPECT().NewMessage("hello").Return(botmsg.MessageFromBot{TextMessageFromBot: botmsg.TextMessageFromBot{Text: "hello"}})
		result := consumeMessaging(mockWHC)
		if result != "hello" {
			t.Errorf("Expected 'hello', got %q", result)
		}
	})

	t.Run("WebhookContext_satisfies_WebhookInputContext", func(t *testing.T) {
		consumeInputCtx := func(ic botsfw.WebhookInputContext) string {
			return ic.GetBotUserID()
		}
		mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
		mockWHC.EXPECT().GetBotUserID().Return("user42")
		result := consumeInputCtx(mockWHC)
		if result != "user42" {
			t.Errorf("Expected 'user42', got %q", result)
		}
	})

	t.Run("WebhookContext_satisfies_WebhookUserData", func(t *testing.T) {
		consumeUserData := func(ud botsfw.WebhookUserData) string {
			return ud.AppUserID()
		}
		mockWHC := mock_botsfw.NewMockWebhookContext(ctrl)
		mockWHC.EXPECT().AppUserID().Return("app-user-1")
		result := consumeUserData(mockWHC)
		if result != "app-user-1" {
			t.Errorf("Expected 'app-user-1', got %q", result)
		}
	})

	t.Run("WebhookTelemetry_mock_standalone", func(t *testing.T) {
		mockTelemetry := mock_botsfw.NewMockWebhookTelemetry(ctrl)
		mockAnalytics := mock_botsfw.NewMockWebhookAnalytics(ctrl)
		mockTelemetry.EXPECT().Analytics().Return(mockAnalytics)
		result := mockTelemetry.Analytics()
		if result != mockAnalytics {
			t.Error("Expected mock analytics from standalone telemetry mock")
		}
	})
}
