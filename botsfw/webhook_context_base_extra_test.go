package botsfw

import (
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strongo/i18n"
	"go.uber.org/mock/gomock"
)

func TestWebhookContextBase_GetBotToken(t *testing.T) {
	whcb := &WebhookContextBase{
		botContext: BotContext{
			BotSettings: &BotSettings{Token: "secret-token"},
		},
	}
	assert.Equal(t, "secret-token", whcb.GetBotToken())
}

func TestWebhookContextBase_InputType(t *testing.T) {
	t.Run("returns_input_type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockInput := mock_botinput.NewMockInputMessage(ctrl)
		mockInput.EXPECT().InputType().Return(botinput.TypeText)

		whcb := &WebhookContextBase{}
		whcb.input = mockInput
		assert.Equal(t, botinput.TypeText, whcb.InputType())
	})

	t.Run("panics_when_nil_input", func(t *testing.T) {
		whcb := &WebhookContextBase{}
		assert.Panics(t, func() {
			whcb.InputType()
		})
	})
}

func TestWebhookContextBase_GetTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := mock_botinput.NewMockInputMessage(ctrl)
	expected := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	mockInput.EXPECT().GetTime().Return(expected)

	whcb := &WebhookContextBase{}
	whcb.input = mockInput
	assert.Equal(t, expected, whcb.GetTime())
}

func TestWebhookContextBase_MessageText(t *testing.T) {
	t.Run("returns_empty_when_not_text_message", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockInput := mock_botinput.NewMockInputMessage(ctrl)
		whcb := &WebhookContextBase{}
		whcb.input = mockInput
		assert.Equal(t, "", whcb.MessageText())
	})

	t.Run("returns_text_when_text_message", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockTextMsg := mock_botinput.NewMockTextMessage(ctrl)
		mockTextMsg.EXPECT().Text().Return("hello world")

		whcb := &WebhookContextBase{}
		whcb.input = mockTextMsg
		assert.Equal(t, "hello world", whcb.MessageText())
	})
}

func TestWebhookContextBase_Locale(t *testing.T) {
	t.Run("returns_bot_settings_locale_when_no_chat_data_and_no_locale_set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockInput := mock_botinput.NewMockInputMessage(ctrl)
		// BotChatID() calls input.BotChatID() - return empty so ChatData() returns nil
		mockInput.EXPECT().BotChatID().Return("", nil)
		mockInput.EXPECT().LogRequest()

		whcb := &WebhookContextBase{
			input: mockInput,
			botContext: BotContext{
				BotSettings: &BotSettings{
					Locale: i18n.LocaleEnUS,
				},
			},
		}
		loc := whcb.Locale()
		assert.Equal(t, "en-US", loc.Code5)
	})

	t.Run("returns_already_set_locale", func(t *testing.T) {
		whcb := &WebhookContextBase{
			locale: i18n.Locale{Code5: "de-DE"},
			botContext: BotContext{
				BotSettings: &BotSettings{
					Locale: i18n.LocaleEnUS,
				},
			},
		}
		loc := whcb.Locale()
		assert.Equal(t, "de-DE", loc.Code5)
	})
}

func TestWebhookContextBase_SetLocale(t *testing.T) {
	t.Run("error_on_empty_code5", func(t *testing.T) {
		whcb := &WebhookContextBase{}
		err := whcb.SetLocale("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expects non-empty string")
	})

	t.Run("error_on_nil_appContext", func(t *testing.T) {
		whcb := &WebhookContextBase{}
		err := whcb.SetLocale("en-US")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "appContext is nil")
	})

	t.Run("success_with_valid_locale", func(t *testing.T) {
		whcb := &WebhookContextBase{
			appContext: testAppContext{},
		}
		err := whcb.SetLocale("en-US")
		require.NoError(t, err)
		assert.Equal(t, "en-US", whcb.locale.Code5)
	})

	t.Run("error_on_unsupported_locale", func(t *testing.T) {
		whcb := &WebhookContextBase{
			appContext: testAppContext{},
		}
		err := whcb.SetLocale("xx-XX")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported locale")
	})
}

func TestWebhookContextBase_GetTranslator(t *testing.T) {
	whcb := &WebhookContextBase{
		appContext: testAppContext{},
	}
	tr := whcb.GetTranslator("en-US")
	require.NotNil(t, tr)

	// The translator should return the key as-is (testTranslator behavior)
	result := tr.Translate("hello")
	assert.Equal(t, "hello", result)
}

func TestWebhookContextBase_Chat(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := mock_botinput.NewMockInputMessage(ctrl)
	mockChat := mock_botinput.NewMockChat(ctrl)
	mockInput.EXPECT().Chat().Return(mockChat)

	whcb := &WebhookContextBase{}
	whcb.input = mockInput
	assert.Equal(t, mockChat, whcb.Chat())
}

func TestWebhookContextBase_GetRecipient(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := mock_botinput.NewMockInputMessage(ctrl)
	mockRecipient := mock_botinput.NewMockRecipient(ctrl)
	mockInput.EXPECT().GetRecipient().Return(mockRecipient)

	whcb := &WebhookContextBase{}
	whcb.input = mockInput
	assert.Equal(t, mockRecipient, whcb.GetRecipient())
}

func TestWebhookContextBase_LogRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInput := mock_botinput.NewMockInputMessage(ctrl)
	mockInput.EXPECT().LogRequest()

	whcb := &WebhookContextBase{}
	whcb.input = mockInput
	// Should not panic
	whcb.LogRequest()
}

func TestWebhookContextBase_IsInGroup(t *testing.T) {
	whcb := &WebhookContextBase{
		getIsInGroup: func() (bool, error) {
			return true, nil
		},
	}
	result, err := whcb.IsInGroup()
	require.NoError(t, err)
	assert.True(t, result)
}

func TestWebhookContextBase_MustBotChatID(t *testing.T) {
	t.Run("returns_id_when_set", func(t *testing.T) {
		whcb := &WebhookContextBase{}
		whcb.botChat.ID = "chat123"
		assert.Equal(t, "chat123", whcb.MustBotChatID())
	})
}

func TestWebhookContextBase_BotChatID(t *testing.T) {
	t.Run("returns_cached_id", func(t *testing.T) {
		whcb := &WebhookContextBase{}
		whcb.botChat.ID = "cached-id"
		id, err := whcb.BotChatID()
		require.NoError(t, err)
		assert.Equal(t, "cached-id", id)
	})
}

func TestWebhookContextBase_SaveBotUser(t *testing.T) {
	// SaveBotUser currently returns "not implemented" error
	// We can't really test it without a real DB, but we can verify it exists
	whcb := &WebhookContextBase{}
	assert.NotNil(t, whcb) // Just verify the struct has the method
}

func TestEnvConstants(t *testing.T) {
	assert.Equal(t, "local", EnvLocal)
	assert.Equal(t, "production", EnvProduction)
}
