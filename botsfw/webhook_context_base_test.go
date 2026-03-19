package botsfw

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebhookContextBase(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("NewWebhookContextBase() did not panic")
			}
		}()
		args := CreateWebhookContextArgs{}
		_, _ = NewWebhookContextBase(args, nil, nil, nil, nil)
	})
}

func TestWebhookContextBase_NewMessage(t *testing.T) {
	whcb := &WebhookContextBase{}
	m := whcb.NewMessage("hello world")
	assert.Equal(t, "hello world", m.Text)
}

func TestWebhookContextBase_NewMessage_empty(t *testing.T) {
	whcb := &WebhookContextBase{}
	m := whcb.NewMessage("")
	assert.Equal(t, "", m.Text)
}

func TestWebhookContextBase_SetUserAndAppUserID(t *testing.T) {
	whcb := &WebhookContextBase{}

	// Before SetUser, appUserID should be empty (assuming no chat data loaded)
	whcb.SetUser("user123", nil)
	assert.Equal(t, "user123", whcb.appUserID)

	// Round-trip: SetUser then read back the appUserID field directly
	whcb.SetUser("user456", nil)
	assert.Equal(t, "user456", whcb.appUserID)
}

func TestWebhookContextBase_ContextRoundTrip(t *testing.T) {
	whcb := &WebhookContextBase{}

	ctx := context.Background()
	whcb.SetContext(ctx)
	assert.Equal(t, ctx, whcb.Context())

	// Replace context with a value context
	type ctxKey string
	ctx2 := context.WithValue(ctx, ctxKey("key"), "val")
	whcb.SetContext(ctx2)
	assert.Equal(t, ctx2, whcb.Context())
	assert.Equal(t, "val", whcb.Context().Value(ctxKey("key")))
}

func TestWebhookContextBase_Environment(t *testing.T) {
	whcb := &WebhookContextBase{
		botContext: BotContext{
			BotSettings: &BotSettings{Env: "production"},
		},
	}
	assert.Equal(t, "production", whcb.Environment())

	whcb.botContext.BotSettings.Env = "local"
	assert.Equal(t, "local", whcb.Environment())
}

func TestWebhookContextBase_BotContext(t *testing.T) {
	settings := &BotSettings{Code: "testbot"}
	whcb := &WebhookContextBase{
		botContext: BotContext{
			BotSettings: settings,
		},
	}
	bc := whcb.BotContext()
	assert.Equal(t, settings, bc.BotSettings)
}

func TestWebhookContextBase_GetBotCode(t *testing.T) {
	whcb := &WebhookContextBase{
		botContext: BotContext{
			BotSettings: &BotSettings{Code: "mybot"},
		},
	}
	assert.Equal(t, "mybot", whcb.GetBotCode())
}

func TestWebhookContextBase_GetBotSettings(t *testing.T) {
	settings := &BotSettings{Code: "mybot", Token: "tok123"}
	whcb := &WebhookContextBase{
		botContext: BotContext{
			BotSettings: settings,
		},
	}
	assert.Equal(t, settings, whcb.GetBotSettings())
}

func TestWebhookContextBase_Request(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.Nil(t, whcb.Request())
}

func TestWebhookContextBase_DB(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.Nil(t, whcb.DB())
}

func TestWebhookContextBase_AppContext(t *testing.T) {
	appCtx := &testAppContext{}
	whcb := &WebhookContextBase{
		appContext: appCtx,
	}
	assert.Equal(t, appCtx, whcb.AppContext())
}

func TestWebhookContextBase_ExecutionContext(t *testing.T) {
	whcb := &WebhookContextBase{}
	ec := whcb.ExecutionContext()
	// ExecutionContext returns whcb itself
	assert.Equal(t, whcb, ec)
}

func TestWebhookContextBase_SetChatID(t *testing.T) {
	whcb := &WebhookContextBase{}
	whcb.SetChatID("chat123")
	assert.Equal(t, "chat123", whcb.botChat.ID)
}

func TestWebhookContextBase_HasChatData(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.False(t, whcb.HasChatData())
}

func TestWebhookContextBase_Analytics(t *testing.T) {
	whcb := &WebhookContextBase{}
	// Analytics is set by NewWebhookContextBase but we can test the accessor
	a := whcb.Analytics()
	// It will be nil-ish (zero value webhookAnalytics) if not properly initialized
	// but should not panic
	_ = a
}

func TestWebhookContextBase_RecordsFieldsSetter(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.Nil(t, whcb.RecordsFieldsSetter())
}

func TestWebhookContextBase_CommandText(t *testing.T) {
	// CommandText calls Translate on the title (if not prefixed with /) and then CommandTextNoTrans
	// We need to set up the translator. Use a minimal setup.
	whcb := &WebhookContextBase{}
	whcb.translator = translator{
		localeCode5: func() string { return "en-US" },
		Translator:  testTranslator{},
	}

	t.Run("title_and_icon", func(t *testing.T) {
		result := whcb.CommandText("mytitle", "🔥")
		// testTranslator returns key as-is, so Translate("mytitle") => "mytitle"
		// Then CommandTextNoTrans("mytitle", "🔥") => "mytitle 🔥"
		assert.Equal(t, "mytitle 🔥", result)
	})

	t.Run("slash_title_not_translated", func(t *testing.T) {
		result := whcb.CommandText("/start", "🚀")
		// Titles starting with "/" are not translated
		assert.Equal(t, "/start 🚀", result)
	})

	t.Run("empty_title_with_icon", func(t *testing.T) {
		result := whcb.CommandText("", "🔥")
		assert.Equal(t, "🔥", result)
	})

	t.Run("title_without_icon", func(t *testing.T) {
		result := whcb.CommandText("hello", "")
		assert.Equal(t, "hello", result)
	})
}

func TestWebhookContextBase_NewMessageByCode(t *testing.T) {
	whcb := &WebhookContextBase{}
	whcb.translator = translator{
		localeCode5: func() string { return "en-US" },
		Translator:  testTranslator{},
	}

	t.Run("without_args", func(t *testing.T) {
		m := whcb.NewMessageByCode("greeting")
		// testTranslator returns key "greeting", then Sprintf("greeting") => "greeting"
		assert.Equal(t, "greeting", m.Text)
	})

	t.Run("with_format_args", func(t *testing.T) {
		m := whcb.NewMessageByCode("Hello %s, you have %d items", "Alice", 5)
		// Translate returns the key "Hello %s, you have %d items"
		// Then Sprintf formats it with args
		assert.Equal(t, "Hello Alice, you have 5 items", m.Text)
	})
}

func TestWebhookContextBase_Input(t *testing.T) {
	whcb := &WebhookContextBase{}
	// Without initialization, input should be nil
	assert.Nil(t, whcb.Input())
}

func TestWebhookContextBase_AppUserEntity(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.Nil(t, whcb.AppUserEntity())
}

func TestWebhookContextBase_BotPlatform(t *testing.T) {
	whcb := &WebhookContextBase{}
	assert.Nil(t, whcb.BotPlatform())
}
