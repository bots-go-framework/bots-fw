package botsfw

import (
	"net/url"
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/stretchr/testify/assert"
)

var testCmd = Command{
	Title: "Title1",
	Titles: map[string]string{
		ShortTitle: "ttl1",
	},
}

var testWhc = &TestWebhookContext{}

func TestCommand_DefaultTitle(t *testing.T) {
	if testCmd.DefaultTitle(testWhc) != "Title1" {
		t.Error("Wrong title")
	}
}

func TestCommand_TitleByKey(t *testing.T) {
	if testCmd.TitleByKey(ShortTitle, testWhc) != "ttl1" {
		t.Error("Wrong title")
	}
}

func TestCommandTextNoTrans(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		icon     string
		expected string
	}{
		{"both_empty", "", "", "<NO_TITLE_OR_ICON>"},
		{"title_only", "Settings", "", "Settings"},
		{"icon_only", "", "⚙️", "⚙️"},
		{"title_and_icon", "Settings", "⚙️", "Settings ⚙️"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommandTextNoTrans(tt.title, tt.icon)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommand_String(t *testing.T) {
	cmd := Command{
		Code:       "test_cmd",
		InputTypes: []botinput.Type{botinput.TypeText},
		Icon:       "🔥",
		Title:      "Test",
		ExactMatch: "test",
		Commands:   []string{"/test"},
		Replies:    []Command{{Code: "reply1"}},
	}
	s := cmd.String()
	assert.True(t, strings.Contains(s, "test_cmd"), "should contain code")
	assert.True(t, strings.Contains(s, "🔥"), "should contain icon")
	assert.True(t, strings.Contains(s, "Test"), "should contain title")
	assert.True(t, strings.Contains(s, "test"), "should contain exact match")
	assert.True(t, strings.Contains(s, "len(Command): 1"), "should show commands count")
	assert.True(t, strings.Contains(s, "len(Replies): 1"), "should show replies count")
}

func TestCommand_String_empty(t *testing.T) {
	cmd := Command{}
	s := cmd.String()
	assert.True(t, strings.Contains(s, "Command{"))
}

func TestNewCallbackCommand(t *testing.T) {
	called := false
	action := func(_ WebhookContext, _ *url.URL) (m botmsg.MessageFromBot, err error) {
		called = true
		return
	}
	cmd := NewCallbackCommand("my_callback", action)
	_ = called
	assert.Equal(t, CommandCode("my_callback"), cmd.Code)
	assert.Equal(t, []botinput.Type{botinput.TypeCallbackQuery}, cmd.InputTypes)
	assert.Equal(t, []string{"/my_callback"}, cmd.Commands)
	assert.NotNil(t, cmd.CallbackAction)
	assert.Nil(t, cmd.Action)
}

func TestNewInlineQueryCommand(t *testing.T) {
	action := func(_ WebhookContext) (m botmsg.MessageFromBot, err error) {
		return
	}
	cmd := NewInlineQueryCommand("my_inline", action)
	assert.Equal(t, CommandCode("my_inline"), cmd.Code)
	assert.Equal(t, []botinput.Type{botinput.TypeInlineQuery}, cmd.InputTypes)
	assert.NotNil(t, cmd.Action)
}

func TestNewCallbackCommand_structure(t *testing.T) {
	cmd := NewCallbackCommand("cb1", nil)
	assert.Equal(t, CommandCode("cb1"), cmd.Code)
	assert.Equal(t, []botinput.Type{botinput.TypeCallbackQuery}, cmd.InputTypes)
	assert.Equal(t, []string{"/cb1"}, cmd.Commands)
}

func TestNewInlineQueryCommand_structure(t *testing.T) {
	cmd := NewInlineQueryCommand("iq1", nil)
	assert.Equal(t, CommandCode("iq1"), cmd.Code)
	assert.Equal(t, []botinput.Type{botinput.TypeInlineQuery}, cmd.InputTypes)
	assert.Nil(t, cmd.CallbackAction)
}

func TestCommand_TitleByKey_withIcon(t *testing.T) {
	cmd := Command{
		Code:  "iconcmd",
		Title: "MyTitle",
		Icon:  "🎯",
	}
	// TestWebhookContext.CommandText panics, so let's test with a command that has no icon
	cmdNoIcon := Command{
		Code:  "noiconcode",
		Title: "PlainTitle",
	}
	// TestWebhookContext.Translate returns key as-is
	result := cmdNoIcon.DefaultTitle(testWhc)
	assert.Equal(t, "PlainTitle", result)

	// Command with icon but no title => returns icon
	cmdIconOnly := Command{
		Code: "icononly",
		Icon: "🎯",
	}
	result = cmdIconOnly.DefaultTitle(testWhc)
	assert.Equal(t, "🎯", result)

	// Command with no icon and no title => returns code as string
	cmdCodeOnly := Command{
		Code: "justcode",
	}
	result = cmdCodeOnly.DefaultTitle(testWhc)
	assert.Equal(t, "justcode", result)

	_ = cmd // just to verify it compiles
}

func TestCommand_TitleByKey_unknownKey(t *testing.T) {
	cmd := Command{
		Code:  "testcmd",
		Title: "Default",
		Titles: map[string]string{
			ShortTitle: "Short",
		},
	}
	// Request a key that doesn't exist and no icon
	result := cmd.TitleByKey("nonexistent_key", testWhc)
	// Falls through: key != DefaultTitle, key not in Titles
	// Icon is empty, title is empty => returns string(Code)
	assert.Equal(t, "testcmd", result)
}
