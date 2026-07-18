package botsdal

import (
	"testing"

	"github.com/bots-go-framework/bots-fw/botsfwconst"
)

func TestNewPlatformKey(t *testing.T) {
	tests := []struct {
		name        string
		platform    botsfwconst.Platform
		shouldPanic bool
		wantID      string
	}{
		{
			name:     "valid_platform",
			platform: "telegram",
			wantID:   "telegram",
		},
		{
			name:     "another_valid_platform",
			platform: "viber",
			wantID:   "viber",
		},
		{
			name:        "empty_platform_panics",
			platform:    "",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewPlatformKey() did not panic with empty platform")
					}
				}()
			}
			got := NewPlatformKey(tt.platform)
			if tt.shouldPanic {
				t.Errorf("NewPlatformKey() should have panicked")
				return
			}
			if got == nil {
				t.Fatal("NewPlatformKey() returned nil")
			}
			if got.Collection() != botPlatformsCollection {
				t.Errorf("NewPlatformKey() collection = %q, want %q", got.Collection(), botPlatformsCollection)
			}
			if got.ID != tt.wantID {
				t.Errorf("NewPlatformKey() ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.Parent() != nil {
				t.Errorf("NewPlatformKey() should have no parent, got %v", got.Parent())
			}
		})
	}
}
