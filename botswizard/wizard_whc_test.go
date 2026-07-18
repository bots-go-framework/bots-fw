package botswizard

// Tests for Command(), Start(), and reply() that require a WebhookContext.
// These use the existing gomock infrastructure already in the module.

import (
	"errors"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botinput"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfw"
	"github.com/bots-go-framework/bots-fw/mocks/mock_botsfwmodels"
	"go.uber.org/mock/gomock"
)

// buildMockWhcWithChatData returns a MockWebhookContext whose ChatData() returns
// the supplied MockBotChatData. Both are pre-wired together.
func buildMockWhcWithChatData(ctrl *gomock.Controller) (*mock_botsfw.MockWebhookContext, *mock_botsfwmodels.MockBotChatData) {
	whc := mock_botsfw.NewMockWebhookContext(ctrl)
	cd := mock_botsfwmodels.NewMockBotChatData(ctrl)
	whc.EXPECT().ChatData().Return(cd).AnyTimes()
	return whc, cd
}

// ---------------------------------------------------------------------------
// reply() – non-complete path (no OnComplete needed)
// ---------------------------------------------------------------------------

func TestWizard_Reply_TextPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc, _ := buildMockWhcWithChatData(ctrl)
	_ = whc // whc is not used in this test; reply only reads it when complete

	w := threeStepWizard()
	r := result{text: "Hello prompt", complete: false}
	m, err := w.reply(whc, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "Hello prompt" {
		t.Errorf("expected 'Hello prompt', got %q", m.Text)
	}
	if m.Format != botmsg.FormatHTML {
		t.Errorf("expected HTML format, got %v", m.Format)
	}
}

// ---------------------------------------------------------------------------
// reply() – complete path calls OnComplete
// ---------------------------------------------------------------------------

func TestWizard_Reply_CompletePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc, _ := buildMockWhcWithChatData(ctrl)

	called := false
	w := threeStepWizard()
	w.OnComplete = func(ctx botsfw.WebhookContext, v Values) (botmsg.MessageFromBot, error) {
		called = true
		if v.String("title") != "myTitle" {
			t.Errorf("OnComplete: wrong title %q", v.String("title"))
		}
		var msg botmsg.MessageFromBot
		msg.Text = "Done!"
		return msg, nil
	}

	r := result{complete: true, values: Values{"title": "myTitle"}}
	m, err := w.reply(whc, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("OnComplete was not called")
	}
	if m.Text != "Done!" {
		t.Errorf("expected 'Done!', got %q", m.Text)
	}
}

// ---------------------------------------------------------------------------
// reply() – complete path propagates OnComplete error
// ---------------------------------------------------------------------------

func TestWizard_Reply_CompletePathError(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc, _ := buildMockWhcWithChatData(ctrl)

	boom := errors.New("db exploded")
	w := threeStepWizard()
	w.OnComplete = func(_ botsfw.WebhookContext, _ Values) (botmsg.MessageFromBot, error) {
		return botmsg.MessageFromBot{}, boom
	}

	r := result{complete: true, values: Values{}}
	_, err := w.reply(whc, r)
	if err != boom {
		t.Errorf("expected boom error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Start() wires whc.ChatData() into start() and calls reply()
// ---------------------------------------------------------------------------

func TestWizard_Start_FirstPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc, cd := buildMockWhcWithChatData(ctrl)

	// start() calls SetAwaitingReplyTo and AddWizardParam(stepParamKey, "0")
	cd.EXPECT().SetAwaitingReplyTo(string("demo")).AnyTimes()
	cd.EXPECT().AddWizardParam(gomock.Any(), gomock.Any()).AnyTimes()
	cd.EXPECT().GetWizardParam(gomock.Any()).Return("").AnyTimes()

	w := threeStepWizard()
	m, err := w.Start(whc, nil)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if m.Text != "Title?" {
		t.Errorf("expected 'Title?', got %q", m.Text)
	}
	if m.Format != botmsg.FormatHTML {
		t.Errorf("expected HTML format, got %v", m.Format)
	}
}

// ---------------------------------------------------------------------------
// Command() – non-TextMessage input returns empty message (ok=false branch)
// ---------------------------------------------------------------------------

func TestWizard_Command_NonTextInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc := mock_botsfw.NewMockWebhookContext(ctrl)

	// MockInputMessage satisfies botinput.InputMessage but NOT botinput.TextMessage,
	// so the type assertion inside Command() returns ok=false and we get empty output.
	nonText := mock_botinput.NewMockInputMessage(ctrl)
	whc.EXPECT().Input().Return(nonText)

	w := threeStepWizard()
	cmd := w.Command()

	m, err := cmd.Action(whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Text != "" {
		t.Errorf("expected empty message for non-text input, got %q", m.Text)
	}
}

// ---------------------------------------------------------------------------
// Command() – TextMessage input drives the wizard through one step
// ---------------------------------------------------------------------------

func TestWizard_Command_TextInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	whc, cd := buildMockWhcWithChatData(ctrl)

	// Wire up the TextMessage mock.
	tm := mock_botinput.NewMockTextMessage(ctrl)
	tm.EXPECT().Text().Return("Run 5k")
	whc.EXPECT().Input().Return(tm)

	// State expectations: handle() reads stepParamKey (returns "" → step 0),
	// then AddWizardParam for the answer, then AddWizardParam for the next step.
	cd.EXPECT().GetWizardParam(tsParamKey).Return("").AnyTimes()
	cd.EXPECT().GetWizardParam(stepParamKey).Return("").AnyTimes()
	cd.EXPECT().GetWizardParam(gomock.Any()).Return("").AnyTimes()
	cd.EXPECT().AddWizardParam(gomock.Any(), gomock.Any()).AnyTimes()

	w := threeStepWizard()
	cmd := w.Command()
	m, err := cmd.Action(whc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After providing title "Run 5k" the next prompt should be "How many?".
	if m.Text != "How many?" {
		t.Errorf("expected 'How many?', got %q", m.Text)
	}
}
