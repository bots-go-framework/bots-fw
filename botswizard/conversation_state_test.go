package botswizard

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestConversationAdapterRoundTripAndCancel(t *testing.T) {
	st := &fakeState{}
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	a := ConversationAdapter{Now: func() time.Time { return now }}
	expires := now.Add(time.Minute)
	if err := a.Save(st, ConversationState{Revision: 4, Feature: "debtus", Flow: "receipt", Step: "channel", Payload: map[string]string{"id": "r1"}, ExpiresAt: expires}); err != nil {
		t.Fatal(err)
	}
	got, ok := a.Load(st)
	if !ok || got.Version != 1 || got.Revision != 4 || got.Feature != "debtus" || got.Flow != "receipt" || got.Step != "channel" || got.Payload["id"] != "r1" {
		t.Fatalf("unexpected state: %#v, %v", got, ok)
	}
	a.Cancel(st)
	if st.GetAwaitingReplyTo() != "" {
		t.Fatal("cancel did not clear legacy state")
	}
}

func TestConversationAdapterExpiresAtClockBoundaryAndClearsState(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	a := ConversationAdapter{Now: func() time.Time { return now }}
	st := &fakeState{}
	if err := a.Save(st, ConversationState{Feature: "debtus", Flow: "receipt", ExpiresAt: now}); err != nil {
		t.Fatal(err)
	}
	_, ok, err := a.LoadChecked(st)
	if ok || !errors.Is(err, ErrConversationExpired) {
		t.Fatalf("load = ok:%v err:%v, want expired", ok, err)
	}
	if st.GetAwaitingReplyTo() != "" {
		t.Fatal("expired conversation was not cleared")
	}
}

func TestConversationAdapterRejectsCorruptAndOversizedState(t *testing.T) {
	st := &fakeState{awaiting: "receipt", params: map[string]string{
		conversationFlowKey:     "receipt",
		conversationFeatureKey:  "debtus",
		conversationRevisionKey: "not-a-number",
	}}
	_, ok, err := (ConversationAdapter{}).LoadChecked(st)
	if ok || !errors.Is(err, ErrConversationCorrupt) {
		t.Fatalf("load = ok:%v err:%v, want corrupt", ok, err)
	}
	if st.GetAwaitingReplyTo() != "" {
		t.Fatal("corrupt conversation was not cleared")
	}
	big := strings.Repeat("x", maxConversationPayloadValueLen+1)
	if err := (ConversationAdapter{}).Save(&fakeState{}, ConversationState{Feature: "debtus", Flow: "receipt", Payload: map[string]string{"x": big}}); err == nil {
		t.Fatal("oversized payload was accepted")
	}
}

func TestConversationAdapterSaveIfRevisionRejectsStaleWriter(t *testing.T) {
	st := &fakeState{}
	a := ConversationAdapter{}
	value := ConversationState{Feature: "debtus", Flow: "receipt"}
	if err := a.SaveIfRevision(st, value, 0); err != nil {
		t.Fatal(err)
	}
	if err := a.SaveIfRevision(st, value, 0); !errors.Is(err, ErrConversationRevisionConflict) {
		t.Fatalf("stale save error = %v", err)
	}
	got, ok := a.Load(st)
	if !ok || got.Revision != 1 {
		t.Fatalf("revision = %#v, %v", got, ok)
	}
}

func TestConversationAdapterMigratesAwaitingReplyTo(t *testing.T) {
	st := &fakeState{awaiting: "legacy-flow"}
	got, ok := (ConversationAdapter{}).Load(st)
	if !ok || got.Version != 0 || got.Flow != "legacy-flow" {
		t.Fatalf("unexpected migration: %#v, %v", got, ok)
	}
	if !IsUniversalCancel(" /cancel ") {
		t.Fatal("cancel command not recognized")
	}
}
