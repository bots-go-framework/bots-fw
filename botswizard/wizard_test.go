package botswizard

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeState mirrors the real chatState semantics the driver relies on: params
// live alongside the awaiting-reply state and are cleared when it is reset.
type fakeState struct {
	awaiting string
	params   map[string]string
}

func (f *fakeState) GetAwaitingReplyTo() string { return f.awaiting }
func (f *fakeState) SetAwaitingReplyTo(v string) {
	f.awaiting = strings.TrimLeft(v, "/")
	if f.awaiting == "" {
		f.params = nil // self-clearing, like the query-string params
	}
}
func (f *fakeState) AddWizardParam(k, v string) {
	if f.params == nil {
		f.params = map[string]string{}
	}
	f.params[k] = v
}
func (f *fakeState) GetWizardParam(k string) string { return f.params[k] }

func threeStepWizard() Wizard {
	return Wizard{
		Code: "demo",
		Steps: []Step{
			{Key: "title", Prompt: "Title?", Parse: func(s string) (string, error) {
				s = strings.TrimSpace(s)
				if s == "" {
					return "", errors.New("title required")
				}
				return s, nil
			}},
			{Key: "count", Prompt: "How many?", Parse: func(s string) (string, error) {
				if strings.TrimSpace(s) == "" || strings.ContainsAny(s, "abc") {
					return "", errors.New("need a number")
				}
				return strings.TrimSpace(s), nil
			}},
			{Key: "note", Prompt: "Note?"}, // nil Parse => stored as-is
		},
	}
}

func TestWizard_HappyPath(t *testing.T) {
	st := &fakeState{}
	w := threeStepWizard()

	r := w.start(st, nil)
	if r.complete || r.text != "Title?" {
		t.Fatalf("start: got %+v", r)
	}
	if r = w.handle(st, "Run 5k"); r.text != "How many?" {
		t.Fatalf("after title: %+v", r)
	}
	if r = w.handle(st, "3"); r.text != "Note?" {
		t.Fatalf("after count: %+v", r)
	}
	r = w.handle(st, "for charity")
	if !r.complete {
		t.Fatalf("expected complete, got %+v", r)
	}
	if r.values.String("title") != "Run 5k" || r.values.Int64("count") != 3 || r.values.String("note") != "for charity" {
		t.Fatalf("values: %+v", r.values)
	}
	if st.awaiting != "" {
		t.Fatalf("state not cleared on completion: %q", st.awaiting)
	}
}

func TestWizard_RepromptOnInvalid(t *testing.T) {
	st := &fakeState{}
	w := threeStepWizard()
	w.start(st, nil)
	r := w.handle(st, "   ") // invalid title
	if r.complete || !strings.Contains(r.text, "title required") || !strings.Contains(r.text, "Title?") {
		t.Fatalf("expected re-prompt, got %+v", r)
	}
	// Still on step 0.
	if got := st.GetWizardParam(stepParamKey); got != "" && got != "0" {
		t.Fatalf("step advanced on invalid input: %q", got)
	}
	if r = w.handle(st, "Valid"); r.text != "How many?" {
		t.Fatalf("recovery failed: %+v", r)
	}
}

func TestWizard_Prefill(t *testing.T) {
	st := &fakeState{}
	w := threeStepWizard()
	r := w.start(st, Values{"title": "Prefilled"})
	if r.complete || r.text != "How many?" {
		t.Fatalf("prefill should skip to step 2: %+v", r)
	}
	// Complete and confirm the prefilled value survives.
	w.handle(st, "5")
	r = w.handle(st, "note")
	if !r.complete || r.values.String("title") != "Prefilled" {
		t.Fatalf("prefilled title lost: %+v", r)
	}
}

func TestWizard_Cancel(t *testing.T) {
	st := &fakeState{}
	w := threeStepWizard()
	w.start(st, nil)
	r := w.handle(st, "/cancel")
	if r.complete || r.text != "Cancelled." || st.awaiting != "" {
		t.Fatalf("cancel failed: %+v awaiting=%q", r, st.awaiting)
	}
}

func TestWizard_Back(t *testing.T) {
	st := &fakeState{}
	w := threeStepWizard()
	w.start(st, nil)
	w.handle(st, "Run 5k") // now on step 1
	r := w.handle(st, "/back")
	if r.text != "Title?" {
		t.Fatalf("back should return to title: %+v", r)
	}
	// /back on the first step cancels.
	r = w.handle(st, "/back")
	if r.text != "Cancelled." || st.awaiting != "" {
		t.Fatalf("back on first step should cancel: %+v", r)
	}
}

func TestWizard_TTLTimeout(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	st := &fakeState{}
	w := threeStepWizard()
	w.TTL = time.Minute
	w.Now = func() time.Time { return now }
	w.start(st, nil)

	// Advance the clock past the TTL before the next reply.
	now = now.Add(2 * time.Minute)
	r := w.handle(st, "Run 5k")
	if !strings.Contains(r.text, "timed out") || st.awaiting != "" {
		t.Fatalf("expected timeout, got %+v", r)
	}
}

func TestValues_TypedAccessors(t *testing.T) {
	v := Values{"n": "42", "when": "2026-08-01T00:00:00Z", "s": "hi"}
	if v.Int64("n") != 42 {
		t.Errorf("Int64: %d", v.Int64("n"))
	}
	if got := v.Time("when", time.RFC3339); got.Year() != 2026 || got.Month() != 8 {
		t.Errorf("Time: %v", got)
	}
	if v.String("s") != "hi" {
		t.Errorf("String: %q", v.String("s"))
	}
	if v.Int64("missing") != 0 {
		t.Errorf("missing Int64 should be 0")
	}
}
