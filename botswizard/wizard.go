// Package botswizard is a thin declarative wrapper over the framework's
// built-in wizard-param machinery (ChatData's AwaitingReplyTo plus
// AddWizardParam/GetWizardParam, where params ride in the AwaitingReplyTo query
// string and self-clear when it is reset).
//
// Instead of hand-rolling one command per step and threading state by hand, you
// describe a wizard as an ordered list of typed Steps. The driver then handles:
//   - routing each reply to the current step,
//   - re-prompting on invalid input (Step.Parse returns an error),
//   - /cancel and /back,
//   - an optional inactivity TTL,
//   - typed access to the collected values,
//   - self-clearing state on completion.
//
// One Wizard registers a single command (Wizard.Command); the current step
// index and all collected values live in the self-clearing wizard params, so
// there is no persistent state to leak.
package botswizard

import (
	"strconv"
	"strings"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
)

// Reserved wizard-param keys (namespaced to avoid clashing with Step.Key).
const (
	stepParamKey = "_wzStep"
	tsParamKey   = "_wzTS"
)

// Step is one question in a wizard.
type Step struct {
	// Key is the wizard-param key the parsed value is stored under and the key
	// it is later read from in Values. Must not start with "_wz".
	Key string
	// Prompt is the question shown when the step is entered (HTML allowed).
	Prompt string
	// Parse validates and normalizes the raw user text into the value to store.
	// Returning an error re-prompts the SAME step with the error text shown
	// above the prompt. A nil Parse stores the trimmed text as-is.
	Parse func(raw string) (stored string, err error)
}

// Values holds the collected step values keyed by Step.Key, with typed getters.
type Values map[string]string

// String returns the raw stored value for key.
func (v Values) String(key string) string { return v[key] }

// Int64 returns the value parsed as int64 (0 if absent/unparseable).
func (v Values) Int64(key string) int64 { n, _ := strconv.ParseInt(v[key], 10, 64); return n }

// Time returns the value parsed with layout (zero time if absent/unparseable).
func (v Values) Time(key, layout string) time.Time { t, _ := time.Parse(layout, v[key]); return t }

// Wizard is a declarative, multi-step, text-input dialog.
type Wizard struct {
	// Code is the command code (also the AwaitingReplyTo path). Unique per bot.
	Code botsfw.CommandCode
	// Steps are asked in order.
	Steps []Step
	// OnComplete runs after the last step with all collected values. Its
	// message is returned to the user. State is already cleared when it runs.
	OnComplete func(whc botsfw.WebhookContext, v Values) (botmsg.MessageFromBot, error)
	// Intro is optional text prepended to the first prompt.
	Intro string
	// TTL, if > 0, abandons the wizard when this long passes since Start.
	TTL time.Duration
	// CancelText / TimeoutText override the default end messages.
	CancelText  string
	TimeoutText string
	// Now overrides the clock for TTL (tests); nil = time.Now.
	Now func() time.Time
}

// state is the subset of ChatData the driver needs (satisfied by whc.ChatData()).
type state interface {
	GetAwaitingReplyTo() string
	SetAwaitingReplyTo(value string)
	AddWizardParam(key, value string)
	GetWizardParam(key string) string
}

// Command returns the single command that drives every step of this wizard.
// Register it with the bot; arm the wizard from your entry command via Start.
func (w Wizard) Command() botsfw.Command {
	return botsfw.Command{
		Code:       w.Code,
		InputTypes: []botinput.Type{botinput.TypeText},
		Action: func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
			tm, ok := whc.Input().(botinput.TextMessage)
			if !ok {
				return botmsg.MessageFromBot{}, nil
			}
			return w.reply(whc, w.handle(whc.ChatData(), tm.Text()))
		},
	}
}

// Start arms the wizard and sends the first prompt. prefill lets an entry
// command supply leading answers (e.g. "/commit run a 5k" prefilling the title);
// each is validated through its Step.Parse and, if valid, that step is skipped.
func (w Wizard) Start(whc botsfw.WebhookContext, prefill Values) (botmsg.MessageFromBot, error) {
	return w.reply(whc, w.start(whc.ChatData(), prefill))
}

// result is the outcome of one state transition.
type result struct {
	text     string
	complete bool
	values   Values
}

func (w Wizard) reply(whc botsfw.WebhookContext, r result) (m botmsg.MessageFromBot, err error) {
	if r.complete {
		return w.OnComplete(whc, r.values)
	}
	// Text/Format are promoted from the embedded TextMessageFromBot, so they are
	// assigned rather than set via a flat struct literal.
	m.Text = r.text
	m.Format = botmsg.FormatHTML
	return m, nil
}

func (w Wizard) start(st state, prefill Values) result {
	st.SetAwaitingReplyTo(string(w.Code))
	if w.TTL > 0 {
		st.AddWizardParam(tsParamKey, w.clock().Format(time.RFC3339))
	}
	idx := 0
	for idx < len(w.Steps) {
		raw, ok := prefill[w.Steps[idx].Key]
		if !ok {
			break
		}
		stored, err := w.parseStep(idx, raw)
		if err != nil {
			break
		}
		st.AddWizardParam(w.Steps[idx].Key, stored)
		idx++
	}
	return w.enter(st, idx, w.Intro)
}

// handle processes one text reply against the current wizard state.
func (w Wizard) handle(st state, text string) result {
	text = strings.TrimSpace(text)

	if isCommand(text, "cancel") {
		return w.cancel(st)
	}
	if w.TTL > 0 {
		if ts := st.GetWizardParam(tsParamKey); ts != "" {
			if started, err := time.Parse(time.RFC3339, ts); err == nil && w.clock().Sub(started) > w.TTL {
				st.SetAwaitingReplyTo("")
				return result{text: w.timeoutText()}
			}
		}
	}

	idx := atoiDefault(st.GetWizardParam(stepParamKey), 0)
	if idx < 0 || idx >= len(w.Steps) { // defensive: corrupt state
		return w.cancel(st)
	}

	if isCommand(text, "back") {
		if idx == 0 {
			return w.cancel(st)
		}
		return w.enter(st, idx-1, "")
	}

	stored, err := w.parseStep(idx, text)
	if err != nil {
		// Re-prompt the same step; state is unchanged.
		return result{text: err.Error() + "\n\n" + w.Steps[idx].Prompt}
	}
	st.AddWizardParam(w.Steps[idx].Key, stored)

	if idx+1 < len(w.Steps) {
		return w.enter(st, idx+1, "")
	}
	// Last step done: collect before clearing (SetAwaitingReplyTo("") wipes params).
	values := w.collect(st)
	st.SetAwaitingReplyTo("")
	return result{complete: true, values: values}
}

// enter positions the wizard at step idx and returns its prompt (with optional
// lead text). If idx is past the end, the wizard completes.
func (w Wizard) enter(st state, idx int, lead string) result {
	if idx >= len(w.Steps) {
		values := w.collect(st)
		st.SetAwaitingReplyTo("")
		return result{complete: true, values: values}
	}
	st.AddWizardParam(stepParamKey, strconv.Itoa(idx))
	text := w.Steps[idx].Prompt
	if lead != "" {
		text = lead + "\n\n" + text
	}
	return result{text: text}
}

func (w Wizard) cancel(st state) result {
	st.SetAwaitingReplyTo("")
	return result{text: w.cancelText()}
}

func (w Wizard) parseStep(idx int, raw string) (string, error) {
	if p := w.Steps[idx].Parse; p != nil {
		return p(raw)
	}
	return strings.TrimSpace(raw), nil
}

func (w Wizard) collect(st state) Values {
	v := make(Values, len(w.Steps))
	for _, s := range w.Steps {
		v[s.Key] = st.GetWizardParam(s.Key)
	}
	return v
}

func (w Wizard) clock() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return time.Now()
}

func (w Wizard) cancelText() string {
	if w.CancelText != "" {
		return w.CancelText
	}
	return "Cancelled."
}

func (w Wizard) timeoutText() string {
	if w.TimeoutText != "" {
		return w.TimeoutText
	}
	return "This step timed out. Please start again."
}

// isCommand reports whether text is the given command word (with/without slash,
// case-insensitive).
func isCommand(text, word string) bool {
	return strings.EqualFold(strings.TrimPrefix(text, "/"), word)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
