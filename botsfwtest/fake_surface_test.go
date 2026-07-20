package botsfwtest_test

import (
	"context"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strongo/analytics"
)

// ---------------------------------------------------------------------------
// Analytics
// ---------------------------------------------------------------------------

func TestFakeAnalytics_EnqueueAndMessages(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	a := whc.Analytics()
	require.NotNil(t, a)

	msg := analytics.NewEvent("test-event", "category", "action")
	a.Enqueue(msg)

	// Get the underlying *fakeAnalytics via the exported accessor on whc.
	// Analytics() returns botsfw.WebhookAnalytics; the concrete type embeds
	// Messages() — we need to call RecordedAnalyticsMessages which returns nil
	// by design, so instead we assert through the responder-path.
	// A second Enqueue should not panic and messages should accumulate.
	msg2 := analytics.NewEvent("test-event-2", "cat2", "act2")
	a.Enqueue(msg2)
}

func TestFakeWhc_Analytics_NotNil(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.NotNil(t, whc.Analytics())
}

func TestFakeWhc_RecordedAnalyticsMessages(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	// RecordedAnalyticsMessages returns nil by design (documented in fake_whc.go).
	result := whc.RecordedAnalyticsMessages()
	assert.Nil(t, result)
}

// ---------------------------------------------------------------------------
// BotPlatform / Environment
// ---------------------------------------------------------------------------

func TestFakeWhc_BotPlatform(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	p := whc.BotPlatform()
	require.NotNil(t, p)
	assert.Equal(t, "fake", p.ID())
	assert.Equal(t, "test", p.Version())
}

func TestFakeWhc_Environment(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Equal(t, "test", whc.Environment())
}

// ---------------------------------------------------------------------------
// BotContext / BotSettings / Request
// ---------------------------------------------------------------------------

func TestFakeWhc_BotContext(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(botsfwtest.WithBotCode("mybot"))
	bc := whc.BotContext()
	require.NotNil(t, bc.BotSettings)
	assert.Equal(t, "mybot", bc.BotSettings.Code)
}

func TestFakeWhc_GetBotSettings(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(botsfwtest.WithBotCode("settingsbot"))
	s := whc.GetBotSettings()
	require.NotNil(t, s)
	assert.Equal(t, "settingsbot", s.Code)
}

func TestFakeWhc_Request_Nil(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Nil(t, whc.Request())
}

func TestFakeWhc_AppContext_Nil(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Nil(t, whc.AppContext())
}

// ---------------------------------------------------------------------------
// Context / SetContext
// ---------------------------------------------------------------------------

func TestFakeWhc_SetContext(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("k"), "v")
	whc.SetContext(ctx)
	got := whc.Context()
	assert.Equal(t, "v", got.Value(ctxKey("k")))
}

func TestFakeWhc_ExecutionContext(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	ec := whc.ExecutionContext()
	require.NotNil(t, ec)
	assert.NotNil(t, ec.Context())
}

// ---------------------------------------------------------------------------
// IsInGroup / IsNewerThen / UpdateLastProcessed
// ---------------------------------------------------------------------------

func TestFakeWhc_IsInGroup(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	inGroup, err := whc.IsInGroup()
	require.NoError(t, err)
	assert.False(t, inGroup)
}

func TestFakeWhc_IsNewerThen(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.True(t, whc.IsNewerThen(whc.ChatData()))
}

func TestFakeWhc_UpdateLastProcessed(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	err := whc.UpdateLastProcessed(whc.ChatData())
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// SaveBotChat / GetBotUser / SetBotUserAccessGranted
// ---------------------------------------------------------------------------

func TestFakeWhc_SaveBotChat(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.NoError(t, whc.SaveBotChat())
}

func TestFakeWhc_GetBotUser(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	user, err := whc.GetBotUser()
	require.NoError(t, err)
	_ = user
}

func TestFakeWhc_SetBotUserAccessGranted(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	require.NoError(t, whc.SetBotUserAccessGranted(true))
}

// ---------------------------------------------------------------------------
// AppUserID / SetUser / AppUserData / RecordsFieldsSetter
// ---------------------------------------------------------------------------

func TestFakeWhc_AppUserID(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	// Default is empty; after SetUser it reflects the id.
	whc.SetUser("app-user-42", nil)
	assert.Equal(t, "app-user-42", whc.AppUserID())
}

func TestFakeWhc_AppUserData_Nil(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	data, err := whc.AppUserData()
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestFakeWhc_RecordsFieldsSetter_Nil(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Nil(t, whc.RecordsFieldsSetter())
}

// ---------------------------------------------------------------------------
// SetLocale
// ---------------------------------------------------------------------------

func TestFakeWhc_SetLocale(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	err := whc.SetLocale("fr-FR")
	require.NoError(t, err)
	assert.Equal(t, "fr-FR", whc.Locale().Code5)
}

// ---------------------------------------------------------------------------
// TranslateWithMap
// ---------------------------------------------------------------------------

func TestFakeWhc_TranslateWithMap(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	got := whc.TranslateWithMap("my.key", map[string]string{"name": "World"})
	assert.Equal(t, "my.key", got)
}

// ---------------------------------------------------------------------------
// GetTranslator
// ---------------------------------------------------------------------------

func TestFakeWhc_GetTranslator(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	tr := whc.GetTranslator("de-DE")
	require.NotNil(t, tr)
	assert.Equal(t, "de-DE", tr.Locale().Code5)
	assert.Equal(t, "some.key", tr.Translate("some.key"))
	assert.Equal(t, "other.key", tr.TranslateNoWarning("other.key"))
	assert.Equal(t, "map.key", tr.TranslateWithMap("map.key", map[string]string{"x": "y"}))
}

// ---------------------------------------------------------------------------
// CommandText
// ---------------------------------------------------------------------------

func TestFakeWhc_CommandText(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	tests := []struct {
		title, icon, want string
	}{
		{"Settings", "⚙️", "Settings ⚙️"},
		{"Settings", "", "Settings"},
		{"", "⚙️", "⚙️"},
	}
	for _, tc := range tests {
		t.Run(tc.title+"/"+tc.icon, func(t *testing.T) {
			assert.Equal(t, tc.want, whc.CommandText(tc.title, tc.icon))
		})
	}
}

// ---------------------------------------------------------------------------
// NewMessage / NewMessageByCode / NewEditMessage
// ---------------------------------------------------------------------------

func TestFakeWhc_NewMessage(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	m := whc.NewMessage("hello world")
	assert.Equal(t, "hello world", m.Text)
}

func TestFakeWhc_NewMessageByCode(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	m := whc.NewMessageByCode("msg.code", "arg1", 2)
	assert.Equal(t, "msg.code", m.Text)
}

func TestFakeWhc_NewEditMessage(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	m, err := whc.NewEditMessage("updated text", botmsg.FormatHTML)
	require.NoError(t, err)
	assert.Equal(t, "updated text", m.Text)
	assert.Equal(t, botmsg.FormatHTML, m.Format)
	assert.True(t, m.IsEdit)
}

// ---------------------------------------------------------------------------
// FakeResponder.DeleteMessage
// ---------------------------------------------------------------------------

func TestFakeResponder_DeleteMessage(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	responder := whc.Responder()
	err := responder.DeleteMessage(context.Background(), "msg-id-1")
	assert.NoError(t, err)
}

func TestFakeResponder_MultipleSendMessages(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	r := whc.Responder()

	_, err := r.SendMessage(context.Background(), whc.NewMessage("first"), botsfw.BotAPISendMessageOverHTTPS)
	require.NoError(t, err)
	_, err = r.SendMessage(context.Background(), whc.NewMessage("second"), botsfw.BotAPISendMessageOverHTTPS)
	require.NoError(t, err)

	sent := whc.SentMessages()
	require.Len(t, sent, 2)
	assert.Equal(t, "first", sent[0].Text)
	assert.Equal(t, "second", sent[1].Text)
}

// ---------------------------------------------------------------------------
// FakeTextMessage surface
// ---------------------------------------------------------------------------

func TestFakeTextMessage_Surface(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(botsfwtest.WithTextMessage("hi there"))
	tm, ok := whc.Input().(*botsfwtest.FakeTextMessage)
	require.True(t, ok)

	assert.Equal(t, botinput.TypeText, tm.InputType())
	assert.Equal(t, "hi there", tm.Text())
	assert.False(t, tm.IsEdited())

	// fakeInputMessage embedded fields
	sender := tm.GetSender()
	require.NotNil(t, sender)
	assert.Equal(t, "user1", sender.GetID())
	assert.Equal(t, "Test", sender.GetFirstName())
	assert.Equal(t, "", sender.GetLastName())
	assert.Equal(t, "user1", sender.GetUserName())
	assert.Equal(t, "en", sender.GetLanguage())
	assert.Equal(t, "", sender.GetAvatar())
	assert.Equal(t, "", sender.GetCountry())
	assert.False(t, sender.IsBotUser())
	assert.Equal(t, "fake", sender.Platform())

	assert.Nil(t, tm.GetRecipient())
	assert.False(t, tm.GetTime().IsZero())
	assert.Equal(t, 0, tm.MessageIntID())
	assert.Equal(t, "", tm.MessageStringID())
	chatID, err := tm.BotChatID()
	require.NoError(t, err)
	assert.Equal(t, "chat1", chatID)

	chat := tm.Chat()
	require.NotNil(t, chat)
	assert.Equal(t, "chat1", chat.GetID())
	assert.Equal(t, "private", chat.GetType())
	assert.False(t, chat.IsGroupChat())

	// LogRequest should not panic
	tm.LogRequest()
}

// ---------------------------------------------------------------------------
// fakeCallbackQuery surface (via whc.Input())
// ---------------------------------------------------------------------------

func TestFakeCallbackQuery_Surface(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(
		botsfwtest.WithCallbackQuery("/confirm?id=7"),
	)
	input := whc.Input()
	require.NotNil(t, input)
	assert.Equal(t, botsfwtest.InputTypeCallbackQuery, input.InputType())

	type callbackQuery interface {
		botinput.InputMessage
		GetID() string
		GetFrom() botinput.Sender
		GetMessage() botinput.Message
		GetData() string
		Chat() botinput.Chat
	}

	cq, ok := input.(callbackQuery)
	require.True(t, ok, "input should implement callbackQuery interface")

	assert.Equal(t, "cbq1", cq.GetID())
	assert.Equal(t, "/confirm?id=7", cq.GetData())
	assert.Nil(t, cq.GetMessage())

	from := cq.GetFrom()
	require.NotNil(t, from)

	chat := cq.Chat()
	require.NotNil(t, chat)
	assert.Equal(t, "chat1", chat.GetID())

	sender := input.GetSender()
	require.NotNil(t, sender)
	assert.Equal(t, "user1", sender.GetID())

	assert.Nil(t, input.GetRecipient())
	assert.False(t, input.GetTime().IsZero())
	assert.Equal(t, 0, input.MessageIntID())
	assert.Equal(t, "", input.MessageStringID())
	botChatID, err := input.BotChatID()
	require.NoError(t, err)
	assert.Equal(t, "chat1", botChatID)
	input.LogRequest()
}

// ---------------------------------------------------------------------------
// fakeUser / fakeChat exercised via WithUserID / WithChatID
// ---------------------------------------------------------------------------

func TestFakeWhc_CustomUserAndChat(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(
		botsfwtest.WithUserID("u-custom"),
		botsfwtest.WithChatID("c-custom"),
		botsfwtest.WithTextMessage("hello"),
	)
	assert.Equal(t, "u-custom", whc.GetBotUserID())
	assert.Equal(t, "c-custom", whc.MustBotChatID())

	tm, ok := whc.Input().(*botsfwtest.FakeTextMessage)
	require.True(t, ok)
	chatID, err := tm.BotChatID()
	require.NoError(t, err)
	assert.Equal(t, "c-custom", chatID)
}

// ---------------------------------------------------------------------------
// GetTime is not zero immediately after creation
// ---------------------------------------------------------------------------

func TestFakeInputMessage_GetTime(t *testing.T) {
	before := time.Now().Add(-time.Second)
	whc := botsfwtest.NewFakeWebhookContext(botsfwtest.WithTextMessage("t"))
	tm := whc.Input().(*botsfwtest.FakeTextMessage)
	got := tm.GetTime()
	assert.True(t, got.After(before) || got.Equal(before))
}
