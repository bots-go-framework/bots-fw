package botsfw

import (
	"context"
	"errors"
	"fmt"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

// ErrSendNotPermitted is the base for refusals returned by a SendGate.
//
// Wrap it so callers can classify a refusal with errors.Is without depending on
// a specific platform's package:
//
//	fmt.Errorf("outside the 24h window: %w", botsfw.ErrSendNotPermitted)
var ErrSendNotPermitted = errors.New("sending is not permitted right now")

// SendGate is an optional interface a WebhookResponder may implement to refuse
// sends its platform does not currently permit.
//
// It exists because "a bot may message any chat it knows, at any time" is a
// Telegram property, not a universal one. Telegram's responder therefore does
// not implement SendGate, and nothing about its behaviour changes.
//
// Other platforms gate sending. WhatsApp permits free-form messages only within
// 24 hours of the recipient's last reply; outside that window a send fails, and
// only a pre-approved template may be delivered. Without this seam a platform has
// no way to say "not now" — the router would send unconditionally, spending an
// API call to earn a rejection, or worse, delivering a billable template message
// the app never intended.
//
// A responder that does not implement SendGate is treated as always permitting,
// so this is additive: existing responders keep working untouched.
type SendGate interface {
	// CanSend reports whether m may be sent right now.
	//
	// A nil error means the send may proceed. A non-nil error means it may not,
	// and describes why; implementations should wrap ErrSendNotPermitted so the
	// refusal is classifiable.
	//
	// CanSend must not perform the send, and should avoid network calls: it is
	// consulted on every outbound message.
	CanSend(c context.Context, m botmsg.MessageFromBot) error
}

// CanSend reports whether responder permits sending m right now.
//
// Responders that do not implement SendGate always permit, so this is safe to
// call on any responder. Returns nil when the send may proceed.
func CanSend(c context.Context, responder WebhookResponder, m botmsg.MessageFromBot) error {
	if responder == nil {
		return nil
	}
	gate, ok := responder.(SendGate)
	if !ok {
		return nil
	}
	return gate.CanSend(c, m)
}

// IsSendNotPermitted reports whether err is a SendGate refusal.
func IsSendNotPermitted(err error) bool {
	return errors.Is(err, ErrSendNotPermitted)
}

// SendMessageThroughGate consults responder's SendGate, if any, and sends only if
// the send is permitted.
//
// This is the single seam every outbound send should route through, so a platform
// gets one place to refuse rather than one per call site. On refusal it returns a
// zero response and the refusal error, having attempted no send.
func SendMessageThroughGate(
	c context.Context,
	responder WebhookResponder,
	m botmsg.MessageFromBot,
	channel botmsg.BotAPISendMessageChannel,
) (OnMessageSentResponse, error) {
	if err := CanSend(c, responder, m); err != nil {
		return OnMessageSentResponse{}, fmt.Errorf("refused by platform send gate: %w", err)
	}
	return responder.SendMessage(c, m, channel)
}
