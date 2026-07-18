package botsdal

import (
	"testing"

	"github.com/bots-go-framework/bots-fw/botsfwconst"
)

func TestNewBotKey(t *testing.T) {
	tests := []struct {
		name        string
		platformID  botsfwconst.Platform
		botID       string
		shouldPanic bool
		wantBotID   string
	}{
		{
			name:       "valid_platform_and_bot",
			platformID: "telegram",
			botID:      "bot123",
			wantBotID:  "bot123",
		},
		{
			name:       "another_platform",
			platformID: "viber",
			botID:      "mybot",
			wantBotID:  "mybot",
		},
		{
			name:        "empty_botID_panics",
			platformID:  "telegram",
			botID:       "",
			shouldPanic: true,
		},
		{
			name:        "empty_platform_panics",
			platformID:  "",
			botID:       "bot123",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewBotKey() did not panic")
					}
				}()
			}
			got := NewBotKey(tt.platformID, tt.botID)
			if tt.shouldPanic {
				t.Errorf("NewBotKey() should have panicked")
				return
			}
			if got == nil {
				t.Fatal("NewBotKey() returned nil")
			}
			if got.Collection() != botsCollection {
				t.Errorf("NewBotKey() collection = %q, want %q", got.Collection(), botsCollection)
			}
			if got.ID != tt.wantBotID {
				t.Errorf("NewBotKey() ID = %v, want %v", got.ID, tt.wantBotID)
			}
			parent := got.Parent()
			if parent == nil {
				t.Fatal("NewBotKey() should have a platform parent key")
			}
			if parent.Collection() != botPlatformsCollection {
				t.Errorf("NewBotKey() parent collection = %q, want %q", parent.Collection(), botPlatformsCollection)
			}
			if parent.ID != string(tt.platformID) {
				t.Errorf("NewBotKey() parent ID = %v, want %v", parent.ID, string(tt.platformID))
			}
		})
	}
}
