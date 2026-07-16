package botsfwtest_test

import (
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoCommand is a trivial command used as the test subject.
var echoCommand = botsfw.Command{
	Code: "echo",
	TextAction: func(whc botsfw.WebhookContext, text string) (botmsg.MessageFromBot, error) {
		return whc.NewMessage("echo: " + text), nil
	},
}

// greetCommand stores a visit counter in chat vars and greets the user.
var greetCommand = botsfw.Command{
	Code: "greet",
	Action: func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
		chat := whc.ChatData()
		count := 0
		if err := botsfwmodels.GetJSONVar(chat, "visits", &count); err != nil {
			return botmsg.MessageFromBot{}, err
		}
		count++
		if err := botsfwmodels.SetJSONVar(chat, "visits", count); err != nil {
			return botmsg.MessageFromBot{}, err
		}
		return whc.NewMessage(whc.Translate("hello")), nil
	},
}

// TestEchoCommand is the primary example: test a command end-to-end with FakeWebhookContext.
func TestEchoCommand(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(
		botsfwtest.WithTextMessage("world"),
		botsfwtest.WithBotCode("mybot"),
	)

	m, err := echoCommand.TextAction(whc, whc.Input().(*botsfwtest.FakeTextMessage).Text())
	require.NoError(t, err)
	assert.Equal(t, "echo: world", m.Text)
}

// TestFakeWebhookContext_defaults verifies that default values are sensible.
func TestFakeWebhookContext_defaults(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Equal(t, "testbot", whc.GetBotCode())
	assert.Equal(t, "chat1", whc.MustBotChatID())
	assert.Equal(t, "user1", whc.GetBotUserID())
	assert.Equal(t, "en-US", whc.Locale().Code5)
	assert.NotNil(t, whc.ChatData())
}

// TestFakeWebhookContext_WithOptions verifies that all options apply correctly.
func TestFakeWebhookContext_WithOptions(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(
		botsfwtest.WithBotCode("specialbot"),
		botsfwtest.WithUserID("u99"),
		botsfwtest.WithChatID("c42"),
		botsfwtest.WithLocale("de-DE"),
		botsfwtest.WithTextMessage("/start"),
	)
	assert.Equal(t, "specialbot", whc.GetBotCode())
	assert.Equal(t, "u99", whc.GetBotUserID())
	assert.Equal(t, "c42", whc.MustBotChatID())
	assert.Equal(t, "de-DE", whc.Locale().Code5)
}

// TestFakeWebhookContext_CallbackQuery verifies callback query input type.
func TestFakeWebhookContext_CallbackQuery(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext(
		botsfwtest.WithCallbackQuery("/confirm?id=42"),
	)
	input := whc.Input()
	assert.Equal(t, botsfwtest.InputTypeCallbackQuery, input.InputType())
}

// TestFakeWebhookContext_ChatVars verifies that GetVar/SetVar work in-memory.
func TestFakeWebhookContext_ChatVars(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	whc.ChatData().SetVar("k", "v")
	assert.Equal(t, "v", whc.ChatData().GetVar("k"))
}

// TestFakeWebhookContext_JSONVars verifies typed JSON vars round-trip correctly.
func TestFakeWebhookContext_JSONVars(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	require.NoError(t, botsfwmodels.SetJSONVar(whc.ChatData(), "count", 7))
	var n int
	require.NoError(t, botsfwmodels.GetJSONVar(whc.ChatData(), "count", &n))
	assert.Equal(t, 7, n)
}

// TestFakeWebhookContext_Responder verifies that SendMessage results are captured.
func TestFakeWebhookContext_Responder(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	m := whc.NewMessage("hello")
	_, err := whc.Responder().SendMessage(whc.Context(), m, botsfw.BotAPISendMessageOverHTTPS)
	require.NoError(t, err)

	sent := whc.SentMessages()
	require.Len(t, sent, 1)
	assert.Equal(t, "hello", sent[0].Text)
}

// TestGreetCommand verifies that a command that writes chat state works correctly.
func TestGreetCommand(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()

	m, err := greetCommand.Action(whc)
	require.NoError(t, err)
	assert.Equal(t, "hello", m.Text) // pass-through translator returns the key

	// Second invocation: counter should be 2.
	m, err = greetCommand.Action(whc)
	require.NoError(t, err)
	assert.Equal(t, "hello", m.Text)

	var count int
	require.NoError(t, botsfwmodels.GetJSONVar(whc.ChatData(), "visits", &count))
	assert.Equal(t, 2, count)
}

// TestFakeWebhookContext_PassThroughTranslator verifies that Translate returns the key.
func TestFakeWebhookContext_PassThroughTranslator(t *testing.T) {
	whc := botsfwtest.NewFakeWebhookContext()
	assert.Equal(t, "some.key", whc.Translate("some.key"))
	assert.Equal(t, "another.key", whc.TranslateNoWarning("another.key"))
}
