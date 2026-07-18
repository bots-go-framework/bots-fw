package botswizard

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// cancelText / timeoutText custom overrides
// ---------------------------------------------------------------------------

func TestWizard_CancelTextCustom(t *testing.T) {
	w := threeStepWizard()
	w.CancelText = "Nope, not happening."
	st := &fakeState{}
	w.start(st, nil)
	r := w.handle(st, "/cancel")
	if r.text != "Nope, not happening." {
		t.Fatalf("expected custom cancel text, got %q", r.text)
	}
}

func TestWizard_TimeoutTextCustom(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	w := threeStepWizard()
	w.TTL = time.Minute
	w.TimeoutText = "Sorry, you were too slow."
	w.Now = func() time.Time { return now }

	st := &fakeState{}
	w.start(st, nil)

	now = now.Add(5 * time.Minute)
	r := w.handle(st, "whatever")
	if r.text != "Sorry, you were too slow." {
		t.Fatalf("expected custom timeout text, got %q", r.text)
	}
	if st.awaiting != "" {
		t.Fatalf("state should be cleared on timeout, got %q", st.awaiting)
	}
}

// ---------------------------------------------------------------------------
// clock() default (real time.Now path)
// ---------------------------------------------------------------------------

func TestWizard_ClockDefault(t *testing.T) {
	w := threeStepWizard()
	// w.Now is nil; clock() must fall through to time.Now without panicking.
	got := w.clock()
	if got.IsZero() {
		t.Fatal("clock() returned zero time")
	}
	// Sanity: it should be in the right decade.
	if got.Year() < 2024 || got.Year() > 2100 {
		t.Fatalf("clock() year looks wrong: %d", got.Year())
	}
}

// ---------------------------------------------------------------------------
// atoiDefault: non-numeric string falls back to default
// ---------------------------------------------------------------------------

func TestAtoiDefault(t *testing.T) {
	tests := []struct {
		s    string
		def  int
		want int
	}{
		{"", 5, 5},     // empty → default
		{"3", 0, 3},    // valid number
		{"abc", 7, 7},  // non-numeric → default
		{"1.5", 0, 0},  // float-ish → default
		{"-2", 99, -2}, // negative valid
	}
	for _, tc := range tests {
		if got := atoiDefault(tc.s, tc.def); got != tc.want {
			t.Errorf("atoiDefault(%q, %d) = %d, want %d", tc.s, tc.def, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// handle: corrupt step param (idx out of range) falls back to cancel
// ---------------------------------------------------------------------------

func TestWizard_CorruptStepParamCancels(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	w.start(st, nil)

	// Manually corrupt the step param to an out-of-range value.
	st.AddWizardParam(stepParamKey, "99")

	r := w.handle(st, "anything")
	if r.complete || r.text != "Cancelled." || st.awaiting != "" {
		t.Fatalf("corrupt state should cancel, got %+v awaiting=%q", r, st.awaiting)
	}
}

func TestWizard_NegativeStepParamCancels(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	w.start(st, nil)

	// Manually set a negative step index.
	st.AddWizardParam(stepParamKey, "-1")

	r := w.handle(st, "anything")
	if r.complete || r.text != "Cancelled." || st.awaiting != "" {
		t.Fatalf("negative step should cancel, got %+v awaiting=%q", r, st.awaiting)
	}
}

// ---------------------------------------------------------------------------
// enter: idx >= len(Steps) triggers completion directly
// ---------------------------------------------------------------------------

func TestWizard_EnterPastEndCompletes(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	w.start(st, nil)

	// Pre-populate all step values so collect() returns them.
	st.AddWizardParam("title", "My Title")
	st.AddWizardParam("count", "7")
	st.AddWizardParam("note", "test note")

	// Call enter() with idx beyond the last step.
	r := w.enter(st, len(w.Steps), "")
	if !r.complete {
		t.Fatalf("expected complete from enter past end, got %+v", r)
	}
	if r.values.String("title") != "My Title" {
		t.Fatalf("values missing, got %+v", r.values)
	}
	if st.awaiting != "" {
		t.Fatalf("state not cleared, awaiting=%q", st.awaiting)
	}
}

// ---------------------------------------------------------------------------
// start: prefill with a failing Parse stops and asks that step
// ---------------------------------------------------------------------------

func TestWizard_PrefillParseError(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}

	// Supply an empty title which the title Parse rejects.
	r := w.start(st, Values{"title": ""})
	// Should land on step 0 (title) since prefill failed.
	if r.complete || r.text != "Title?" {
		t.Fatalf("expected to stop at step 0 on parse error, got %+v", r)
	}
	if st.GetWizardParam(stepParamKey) != "0" {
		t.Fatalf("step should be 0, got %q", st.GetWizardParam(stepParamKey))
	}
}

func TestWizard_PrefillPartialParseError(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}

	// Title is valid, count is invalid. Should stop at step 1 (count).
	r := w.start(st, Values{"title": "Good Title", "count": "abc"})
	if r.complete || r.text != "How many?" {
		t.Fatalf("expected to stop at step 1 on parse error, got %+v", r)
	}
}

// ---------------------------------------------------------------------------
// enter with a lead text
// ---------------------------------------------------------------------------

func TestWizard_EnterWithLead(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	r := w.enter(st, 0, "Welcome!")
	if !strings.Contains(r.text, "Welcome!") || !strings.Contains(r.text, "Title?") {
		t.Fatalf("expected intro+prompt, got %q", r.text)
	}
}

// ---------------------------------------------------------------------------
// /back on step 0 cancels (already covered in TestWizard_Back but added
// explicitly here for clarity alongside the custom-canceltext test)
// ---------------------------------------------------------------------------

func TestWizard_BackAtStepZeroCancels(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	w.start(st, nil)

	r := w.handle(st, "back") // without slash — isCommand handles both forms
	if r.text != "Cancelled." || st.awaiting != "" {
		t.Fatalf("back at step 0 should cancel, got %+v", r)
	}
}

// ---------------------------------------------------------------------------
// wizard with a single step: handle completes immediately on valid input
// ---------------------------------------------------------------------------

func TestWizard_SingleStep(t *testing.T) {
	w := Wizard{
		Code: "single",
		Steps: []Step{
			{Key: "name", Prompt: "Your name?"},
		},
	}
	st := &fakeState{}
	r := w.start(st, nil)
	if r.text != "Your name?" {
		t.Fatalf("unexpected prompt: %q", r.text)
	}
	r = w.handle(st, "  Alice  ")
	if !r.complete {
		t.Fatalf("single step should complete after one answer, got %+v", r)
	}
	if r.values.String("name") != "Alice" {
		t.Fatalf("expected trimmed 'Alice', got %q", r.values.String("name"))
	}
}

// ---------------------------------------------------------------------------
// Intro is prepended to first prompt
// ---------------------------------------------------------------------------

func TestWizard_IntroText(t *testing.T) {
	w := threeStepWizard()
	w.Intro = "Let's get started."
	st := &fakeState{}
	r := w.start(st, nil)
	if !strings.HasPrefix(r.text, "Let's get started.") {
		t.Fatalf("intro not prepended: %q", r.text)
	}
	if !strings.Contains(r.text, "Title?") {
		t.Fatalf("prompt missing from intro text: %q", r.text)
	}
}

// ---------------------------------------------------------------------------
// OnComplete integration via handle (exercises the complete→values path)
// ---------------------------------------------------------------------------

func TestWizard_OnComplete(t *testing.T) {
	var capturedValues Values
	w := threeStepWizard()
	// OnComplete callback capture (not called via reply() since we test handle directly)
	// but we can test collect+complete result by inspecting r.complete and r.values.
	_ = capturedValues

	st := &fakeState{}
	w.start(st, nil)
	w.handle(st, "Title")
	w.handle(st, "5")
	r := w.handle(st, "my note")

	if !r.complete {
		t.Fatalf("expected complete=true, got %+v", r)
	}
	if r.values.String("title") != "Title" || r.values.String("count") != "5" || r.values.String("note") != "my note" {
		t.Fatalf("unexpected values: %+v", r.values)
	}
}

// ---------------------------------------------------------------------------
// cancel without wizard param set (edge: SetAwaitingReplyTo clears params)
// ---------------------------------------------------------------------------

func TestWizard_CancelDefaultText(t *testing.T) {
	w := threeStepWizard()
	// No CancelText set — should use default.
	if w.cancelText() != "Cancelled." {
		t.Errorf("default cancel text wrong: %q", w.cancelText())
	}
}

func TestWizard_TimeoutTextDefault(t *testing.T) {
	w := threeStepWizard()
	if w.timeoutText() != "This step timed out. Please start again." {
		t.Errorf("default timeout text wrong: %q", w.timeoutText())
	}
}

// ---------------------------------------------------------------------------
// isCommand: various forms
// ---------------------------------------------------------------------------

func TestIsCommand(t *testing.T) {
	tests := []struct {
		text string
		word string
		want bool
	}{
		{"/cancel", "cancel", true},
		{"cancel", "cancel", true},
		{"CANCEL", "cancel", true},
		{"/BACK", "back", true},
		{"/start", "cancel", false},
		{"", "cancel", false},
	}
	for _, tc := range tests {
		if got := isCommand(tc.text, tc.word); got != tc.want {
			t.Errorf("isCommand(%q, %q) = %v, want %v", tc.text, tc.word, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Values.Time with missing or unparseable key
// ---------------------------------------------------------------------------

func TestValues_TimeZeroOnMissing(t *testing.T) {
	v := Values{}
	if !v.Time("missing", time.RFC3339).IsZero() {
		t.Error("expected zero time for missing key")
	}
	v["bad"] = "not-a-time"
	if !v.Time("bad", time.RFC3339).IsZero() {
		t.Error("expected zero time for unparseable value")
	}
}

// ---------------------------------------------------------------------------
// prefill skipping all steps → completes immediately
// ---------------------------------------------------------------------------

func TestWizard_PrefillAll(t *testing.T) {
	w := threeStepWizard()
	st := &fakeState{}
	r := w.start(st, Values{
		"title": "T",
		"count": "3",
		"note":  "n",
	})
	if !r.complete {
		t.Fatalf("all steps prefilled should complete immediately, got %+v", r)
	}
	if r.values.String("title") != "T" || r.values.String("count") != "3" || r.values.String("note") != "n" {
		t.Fatalf("wrong values: %+v", r.values)
	}
}

// ---------------------------------------------------------------------------
// TTL: no timeout when within TTL window
// ---------------------------------------------------------------------------

func TestWizard_TTLNoTimeoutWithinWindow(t *testing.T) {
	now := time.Unix(3_000_000, 0)
	w := threeStepWizard()
	w.TTL = 10 * time.Minute
	w.Now = func() time.Time { return now }
	st := &fakeState{}
	w.start(st, nil)

	// Only 30 seconds have passed — well within TTL.
	now = now.Add(30 * time.Second)
	r := w.handle(st, "My Title")
	if strings.Contains(r.text, "timed out") {
		t.Fatalf("should NOT timeout within TTL, got %q", r.text)
	}
	if r.text != "How many?" {
		t.Fatalf("expected step 2 prompt, got %q", r.text)
	}
}

// ---------------------------------------------------------------------------
// handle: re-prompt includes error message above the prompt
// ---------------------------------------------------------------------------

func TestWizard_RepromptFormat(t *testing.T) {
	w := Wizard{
		Code: "rp",
		Steps: []Step{
			{Key: "email", Prompt: "Email?", Parse: func(s string) (string, error) {
				if !strings.Contains(s, "@") {
					return "", errors.New("not a valid email")
				}
				return s, nil
			}},
		},
	}
	st := &fakeState{}
	w.start(st, nil)
	r := w.handle(st, "notanemail")
	if !strings.HasPrefix(r.text, "not a valid email") {
		t.Fatalf("error should come first in re-prompt, got %q", r.text)
	}
	if !strings.HasSuffix(r.text, "Email?") {
		t.Fatalf("prompt should come last in re-prompt, got %q", r.text)
	}
}
