package botsfw

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-go-core/botkb"
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
	if err := (PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}).Validate(m, false); err == nil {
		t.Fatal("expected host-only refusal")
	}
	if err := (PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}).Validate(m, true); err != nil {
		t.Fatalf("host keyboard: %v", err)
	}
}

func TestHostOnlyResponderOwnershipCannotBeChosenByFeatureMessage(t *testing.T) {
	m := botmsg.MessageFromBot{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}
	policy := PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}
	featureNext := &presentationTestResponder{}
	if _, err := NewPolicyResponder(featureNext, policy).SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err == nil {
		t.Fatal("feature-owned responder sent a host-only keyboard")
	}
	hostNext := &presentationTestResponder{}
	if _, err := NewHostPolicyResponder(hostNext, policy).SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err != nil {
		t.Fatalf("host-owned responder: %v", err)
	}
	if hostNext.sends != 1 {
		t.Fatalf("host sends = %d, want 1", hostNext.sends)
	}
}

func TestPresentationPolicyCannotBypassWithUnmarkedBottomKeyboard(t *testing.T) {
	next := &presentationTestResponder{}
	responder := NewPolicyResponder(next, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardDeny})
	m := botmsg.MessageFromBot{}
	m.Keyboard = botkb.NewMessageKeyboard(botkb.KeyboardTypeBottom)
	if _, err := responder.SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err == nil {
		t.Fatal("unmarked bottom keyboard bypassed policy")
	}
	if next.sends != 0 {
		t.Fatal("bypassed policy reached responder")
	}
}

func TestPresentationPolicyAcceptsTypedNilKeyboard(t *testing.T) {
	next := &presentationTestResponder{}
	responder := NewPolicyResponder(next, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardDeny})
	var keyboard *botkb.MessageKeyboard
	m := botmsg.MessageFromBot{}
	m.Keyboard = keyboard
	if _, err := responder.SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err != nil {
		t.Fatalf("typed nil keyboard: %v", err)
	}
	if next.sends != 1 {
		t.Fatalf("delegate sends = %d, want 1", next.sends)
	}
}

func TestCommandPolicyResponderUsesRouterOwnedHostCommandOnly(t *testing.T) {
	next := &presentationTestResponder{}
	responder := NewCommandPolicyResponder(next, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}, "main_menu")
	m := botmsg.MessageFromBot{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}
	if _, err := responder.SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err == nil {
		t.Fatal("direct feature send was granted host authority")
	}
	if _, err := ResponseResponderForCommand(responder, "debtus").SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err == nil {
		t.Fatal("embedded feature command was granted host authority")
	}
	if _, err := ResponseResponderForCommand(responder, "main_menu").SendMessage(context.Background(), m, BotAPISendMessageOverHTTPS); err != nil {
		t.Fatalf("host router response: %v", err)
	}
	if next.sends != 1 {
		t.Fatalf("host sends = %d, want 1", next.sends)
	}
}
