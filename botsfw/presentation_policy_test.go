package botsfw

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

type presentationTestResponder struct{ sends int }

func (r *presentationTestResponder) SendMessage(context.Context, botmsg.MessageFromBot, botmsg.BotAPISendMessageChannel) (OnMessageSentResponse, error) {
	r.sends++
	return OnMessageSentResponse{}, nil
}
func (*presentationTestResponder) DeleteMessage(context.Context, string) error { return nil }

func TestPresentationPolicyResponderAppliesToDirectSends(t *testing.T) {
	next := &presentationTestResponder{}
	responder := NewPolicyResponder(next, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardDeny})
	m := botmsg.MessageFromBot{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}
	if _, err := responder.SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err == nil {
		t.Fatal("expected persistent keyboard refusal")
	}
	if next.sends != 0 {
		t.Fatalf("denied send reached delegate %d times", next.sends)
	}
}

func TestPresentationPolicyHostOnly(t *testing.T) {
	m := botmsg.MessageFromBot{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}
	if err := (PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}).Validate(m); err == nil {
		t.Fatal("expected host-only refusal")
	}
	if err := (PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly, HostMayShowBottomKeyboard: true}).Validate(m); err != nil {
		t.Fatalf("host keyboard: %v", err)
	}
}
