package botswebhook

import "context"

// RouteTraceEvent describes the final command-routing decision for one text
// message: which command won and by which rule (or that nothing matched). It is
// emitted by the message router when RoutingTracer is set.
type RouteTraceEvent struct {
	// InputText is the message text that was being routed.
	InputText string
	// AwaitingReplyTo is the chat's AwaitingReplyTo at match time ("" if none).
	AwaitingReplyTo string
	// CommandCode is the command that matched, or "" for a no-match.
	CommandCode string
	// Reason is HOW the command matched, one of:
	//   "command-code"           — explicit "/code"
	//   "start-command"          — "/start <code>"
	//   "command-alias"          — one of Command.Commands
	//   "awaiting-reply"         — the chat is awaiting a reply to this command
	//   "awaiting-reply-replies" — matched a nested reply under the awaiting command
	//   "exact-match"            — Command.ExactMatch
	//   "default-title"          — Command.DefaultTitle
	//   "matcher"                — Command.Matcher returned true
	//   "no-match"               — nothing matched (CommandCode is "")
	Reason string
}

// RoutingTracer, when non-nil, receives a RouteTraceEvent for the final routing
// decision of each text message — which command won and why (or that nothing
// matched). It is an opt-in troubleshooting hook: leave it nil (the default) in
// production for zero overhead, and set it to answer "why did this reply route
// there?" without turning on framework-wide debug logging.
//
// This is a package-level toggle shared by every router; set it from one place
// while debugging. The callback must be cheap and non-blocking.
var RoutingTracer func(ctx context.Context, event RouteTraceEvent)

// traceRoute emits a RouteTraceEvent when RoutingTracer is set (else a no-op).
func traceRoute(ctx context.Context, commandCode, reason, awaitingReplyTo, inputText string) {
	if RoutingTracer != nil {
		RoutingTracer(ctx, RouteTraceEvent{
			InputText:       inputText,
			AwaitingReplyTo: awaitingReplyTo,
			CommandCode:     commandCode,
			Reason:          reason,
		})
	}
}
