package botswebhook

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw/botsfw"
	"go.uber.org/mock/gomock"
)

// catchAllMatcher matches any input — the shape of the conversational-runtime
// text command that regressed the wizard flows.
func catchAllMatcher(_ botsfw.Command, _ botsfw.WebhookContext) bool { return true }

// TestPrecedence_AwaitingReplyBeatsCatchAllMatcher is the regression test for the
// reported bug: while a chat is awaiting a reply (a wizard step), a catch-all
// Matcher command must NOT hijack the reply. This FAILS on the old interleaved
// matcher (which returned the catch-all regardless of registration order) and
// PASSES with the tiered precedence (AwaitingReplyTo outranks Matcher).
func TestPrecedence_AwaitingReplyBeatsCatchAllMatcher(t *testing.T) {
	wizard := botsfw.Command{Code: "commit_wizard", TextAction: dummyTextAction}
	convo := botsfw.Command{Code: "convo", Matcher: catchAllMatcher, TextAction: dummyTextAction}

	// Order must not matter: the awaiting command wins whether registered before
	// or after the catch-all.
	orders := map[string][]botsfw.Command{
		"catch-all first": {convo, wizard},
		"catch-all last":  {wizard, convo},
	}
	for name, commands := range orders {
		t.Run(name, func(t *testing.T) {
			router := NewWebhookRouter(nil).(*webhooksRouter)
			ctrl := gomock.NewController(t)
			whc := setupMessageWHC(ctrl, "commit_wizard") // a wizard step is in progress

			matched := router.matchMessageCommands(whc, nil, false, "30 pushups per day", "", commands)
			if matched == nil {
				t.Fatal("expected a match")
			}
			if matched.Code != "commit_wizard" {
				t.Errorf("awaiting reply was hijacked: got %q, want %q", matched.Code, "commit_wizard")
			}
		})
	}
}

// TestPrecedence_AwaitingReplyBeatsExactMatch: an in-progress flow also outranks
// ExactMatch (e.g. a reply that happens to equal a button label). Fails on the
// old matcher (ExactMatch returned first), passes with the tiered precedence.
func TestPrecedence_AwaitingReplyBeatsExactMatch(t *testing.T) {
	wizard := botsfw.Command{Code: "commit_wizard", TextAction: dummyTextAction}
	other := botsfw.Command{Code: "other", ExactMatch: "hello", TextAction: dummyTextAction}

	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	whc := setupMessageWHC(ctrl, "commit_wizard")

	matched := router.matchMessageCommands(whc, nil, false, "hello", "", []botsfw.Command{other, wizard})
	if matched == nil || matched.Code != "commit_wizard" {
		t.Fatalf("awaiting reply must outrank ExactMatch: got %v", matched)
	}
}

// TestPrecedence_MatcherStillMatchesWhenIdle guards that the catch-all is not
// disabled: with no AwaitingReplyTo, the Matcher command still claims free text.
func TestPrecedence_MatcherStillMatchesWhenIdle(t *testing.T) {
	convo := botsfw.Command{Code: "convo", Matcher: catchAllMatcher, TextAction: dummyTextAction}

	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	whc := setupMessageWHC(ctrl, "") // idle

	matched := router.matchMessageCommands(whc, nil, false, "30 pushups per day", "", []botsfw.Command{convo})
	if matched == nil || matched.Code != "convo" {
		t.Fatalf("catch-all Matcher must claim free text when idle: got %v", matched)
	}
}

// TestPrecedence_ExplicitCommandBeatsAwaitingReply: an explicit "/cancel" while a
// wizard is armed still routes to the explicit command (Tier 1 > Tier 2), so a
// user can always escape a flow.
func TestPrecedence_ExplicitCommandBeatsAwaitingReply(t *testing.T) {
	wizard := botsfw.Command{Code: "commit_wizard", TextAction: dummyTextAction}
	cancel := botsfw.Command{Code: "cancel", TextAction: dummyTextAction}

	router := NewWebhookRouter(nil).(*webhooksRouter)
	ctrl := gomock.NewController(t)
	// isCommandText=true → awaitingReplyTo is not consulted; "/cancel" matches Tier 1.
	whc := setupMessageWHC(ctrl, "commit_wizard")

	matched := router.matchMessageCommands(whc, nil, true, "/cancel", "", []botsfw.Command{wizard, cancel})
	if matched == nil || matched.Code != "cancel" {
		t.Fatalf("explicit command must outrank an awaiting reply: got %v", matched)
	}
}

// TestRoutingTracer_EmitsDecision covers the opt-in troubleshooting hook: the
// tracer must report which command won and by which rule (and no-match).
func TestRoutingTracer_EmitsDecision(t *testing.T) {
	var events []RouteTraceEvent
	RoutingTracer = func(_ context.Context, e RouteTraceEvent) { events = append(events, e) }
	t.Cleanup(func() { RoutingTracer = nil })

	wizard := botsfw.Command{Code: "commit_wizard", TextAction: dummyTextAction}
	convo := botsfw.Command{Code: "convo", Matcher: catchAllMatcher, TextAction: dummyTextAction}
	plain := botsfw.Command{Code: "plain", TextAction: dummyTextAction}

	cases := []struct {
		name       string
		awaiting   string
		commands   []botsfw.Command
		text       string
		wantCode   string
		wantReason string
	}{
		{"awaiting wins", "commit_wizard", []botsfw.Command{convo, wizard}, "hi", "commit_wizard", "awaiting-reply"},
		{"matcher when idle", "", []botsfw.Command{convo}, "hi", "convo", "matcher"},
		{"no match", "", []botsfw.Command{plain}, "hi", "", "no-match"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			events = nil
			router := NewWebhookRouter(nil).(*webhooksRouter)
			ctrl := gomock.NewController(t)
			whc := setupMessageWHC(ctrl, tc.awaiting)

			router.matchMessageCommands(whc, nil, false, tc.text, "", tc.commands)

			if len(events) == 0 {
				t.Fatal("expected a RouteTraceEvent")
			}
			last := events[len(events)-1]
			if last.CommandCode != tc.wantCode || last.Reason != tc.wantReason {
				t.Errorf("trace = {code:%q reason:%q}, want {code:%q reason:%q}", last.CommandCode, last.Reason, tc.wantCode, tc.wantReason)
			}
			if last.AwaitingReplyTo != tc.awaiting || last.InputText != tc.text {
				t.Errorf("trace context = {awaiting:%q text:%q}, want {awaiting:%q text:%q}", last.AwaitingReplyTo, last.InputText, tc.awaiting, tc.text)
			}
		})
	}
}
