package botswebhook

import (
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw/botinput"
)

type testWebhookEntry struct{ id any }

func (e testWebhookEntry) GetID() any       { return e.id }
func (testWebhookEntry) GetTime() time.Time { return time.Time{} }

type testDurableWebhookEntry struct {
	testWebhookEntry
	updateID string
	ok       bool
}

func (e testDurableWebhookEntry) WebhookUpdateID() (string, bool) { return e.updateID, e.ok }

var _ botinput.DurableWebhookEntry = testDurableWebhookEntry{}

func TestWebhookUpdateID(t *testing.T) {
	tests := []struct {
		name  string
		entry botinput.Entry
		id    string
		ok    bool
	}{
		{name: "nil"},
		{name: "nil ID", entry: testWebhookEntry{}},
		{name: "empty ID", entry: testWebhookEntry{id: "  "}},
		{name: "zero int ID", entry: testWebhookEntry{id: 0}},
		{name: "nonzero compatibility ID", entry: testWebhookEntry{id: 42}, id: "42", ok: true},
		{name: "explicit durable ID", entry: testDurableWebhookEntry{updateID: "update-42", ok: true}, id: "update-42", ok: true},
		{name: "explicit unavailable durable ID", entry: testDurableWebhookEntry{updateID: "update-42", ok: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := webhookUpdateID(tt.entry)
			if id != tt.id || ok != tt.ok {
				t.Fatalf("webhookUpdateID() = %q, %t; want %q, %t", id, ok, tt.id, tt.ok)
			}
		})
	}
}
