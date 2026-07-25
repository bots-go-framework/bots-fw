package botswizard

import (
	"testing"
	"time"
)

func TestConversationAdapterRoundTripAndCancel(t *testing.T) {
	st := &fakeState{}
	a := ConversationAdapter{}
	expires := time.Now().Add(time.Minute).UTC().Truncate(time.Second)
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
