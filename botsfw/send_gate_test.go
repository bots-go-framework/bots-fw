package botsfw

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

// ungatedResponder does not implement SendGate — the Telegram shape.
type ungatedResponder struct {
	sent int
}

func (r *ungatedResponder) SendMessage(
	_ context.Context, _ botmsg.MessageFromBot, _ botmsg.BotAPISendMessageChannel,
) (OnMessageSentResponse, error) {
	r.sent++
	return OnMessageSentResponse{}, nil
}

func (r *ungatedResponder) DeleteMessage(_ context.Context, _ string) error { return nil }

// gatedResponder implements SendGate — the WhatsApp shape.
type gatedResponder struct {
	ungatedResponder
	refuse error
	asked  int
}

func (r *gatedResponder) CanSend(_ context.Context, _ botmsg.MessageFromBot) error {
	r.asked++
	return r.refuse
}

var (
	_ WebhookResponder = (*ungatedResponder)(nil)
	_ WebhookResponder = (*gatedResponder)(nil)
	_ SendGate         = (*gatedResponder)(nil)
)

// TestCanSend_ungatedResponderAlwaysPermits pins the compatibility guarantee:
// a responder that does not implement SendGate is unaffected by this seam.
func TestCanSend_ungatedResponderAlwaysPermits(t *testing.T) {
	if err := CanSend(context.Background(), &ungatedResponder{}, botmsg.MessageFromBot{}); err != nil {
		t.Errorf("an ungated responder must always permit, got: %v", err)
	}
}

func TestCanSend_nilResponderPermits(t *testing.T) {
	if err := CanSend(context.Background(), nil, botmsg.MessageFromBot{}); err != nil {
		t.Errorf("a nil responder must not error, got: %v", err)
	}
}

func TestCanSend_gatePermits(t *testing.T) {
	r := &gatedResponder{}
	if err := CanSend(context.Background(), r, botmsg.MessageFromBot{}); err != nil {
		t.Errorf("expected permit, got: %v", err)
	}
	if r.asked != 1 {
		t.Errorf("expected the gate to be consulted once, got %d", r.asked)
	}
}

func TestCanSend_gateRefuses(t *testing.T) {
	want := fmt.Errorf("outside the 24h window: %w", ErrSendNotPermitted)
	r := &gatedResponder{refuse: want}

	err := CanSend(context.Background(), r, botmsg.MessageFromBot{})
	if err == nil {
		t.Fatal("expected a refusal")
	}
	if !IsSendNotPermitted(err) {
		t.Errorf("refusal must be classifiable via IsSendNotPermitted, got: %v", err)
	}
}

// TestSendMessageThroughGate_refusalSendsNothing is the point of the whole seam:
// a refused send must cost no API call. On WhatsApp an attempted out-of-window
// send earns a rejection, and an attempted template send costs real money.
func TestSendMessageThroughGate_refusalSendsNothing(t *testing.T) {
	r := &gatedResponder{refuse: fmt.Errorf("outside the 24h window: %w", ErrSendNotPermitted)}

	_, err := SendMessageThroughGate(
		context.Background(), r, botmsg.MessageFromBot{}, BotAPISendMessageOverHTTPS,
	)
	if err == nil {
		t.Fatal("expected a refusal")
	}
	if !IsSendNotPermitted(err) {
		t.Errorf("refusal must survive wrapping, got: %v", err)
	}
	if r.sent != 0 {
		t.Errorf("a refused send must not reach the platform, but SendMessage ran %d time(s)", r.sent)
	}
}

func TestSendMessageThroughGate_permittedSendProceeds(t *testing.T) {
	r := &gatedResponder{}
	if _, err := SendMessageThroughGate(
		context.Background(), r, botmsg.MessageFromBot{}, BotAPISendMessageOverHTTPS,
	); err != nil {
		t.Fatalf("expected the send to proceed, got: %v", err)
	}
	if r.sent != 1 {
		t.Errorf("expected exactly 1 send, got %d", r.sent)
	}
}

// TestSendMessageThroughGate_ungatedResponderProceeds pins that existing
// responders keep sending exactly as before.
func TestSendMessageThroughGate_ungatedResponderProceeds(t *testing.T) {
	r := &ungatedResponder{}
	if _, err := SendMessageThroughGate(
		context.Background(), r, botmsg.MessageFromBot{}, BotAPISendMessageOverHTTPS,
	); err != nil {
		t.Fatalf("expected the send to proceed, got: %v", err)
	}
	if r.sent != 1 {
		t.Errorf("expected exactly 1 send, got %d", r.sent)
	}
}

// TestIsSendNotPermitted_unrelatedError pins that ordinary send failures are not
// mistaken for refusals.
func TestIsSendNotPermitted_unrelatedError(t *testing.T) {
	if IsSendNotPermitted(errors.New("connection reset")) {
		t.Error("an unrelated error must not classify as a refusal")
	}
	if IsSendNotPermitted(nil) {
		t.Error("nil must not classify as a refusal")
	}
}
