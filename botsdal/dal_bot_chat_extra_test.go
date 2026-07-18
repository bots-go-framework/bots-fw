package botsdal

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
)

func TestNewBotChatKey(t *testing.T) {
	tests := []struct {
		name        string
		platformID  botsfwconst.Platform
		botID       string
		chatID      string
		shouldPanic bool
		wantChatID  interface{}
	}{
		{
			name:       "valid_key",
			platformID: "telegram",
			botID:      "bot1",
			chatID:     "chat42",
			wantChatID: "chat42",
		},
		{
			name:        "empty_botID_panics",
			platformID:  "telegram",
			botID:       "",
			chatID:      "chat42",
			shouldPanic: true,
		},
		{
			name:        "empty_platform_panics",
			platformID:  "",
			botID:       "bot1",
			chatID:      "chat42",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewBotChatKey() did not panic")
					}
				}()
			}
			got := NewBotChatKey(tt.platformID, tt.botID, tt.chatID)
			if tt.shouldPanic {
				t.Errorf("NewBotChatKey() should have panicked")
				return
			}
			if got == nil {
				t.Fatal("NewBotChatKey() returned nil")
			}
			if got.Collection() != botChatsCollection {
				t.Errorf("NewBotChatKey() collection = %q, want %q", got.Collection(), botChatsCollection)
			}
			if got.ID != tt.wantChatID {
				t.Errorf("NewBotChatKey() ID = %v, want %v", got.ID, tt.wantChatID)
			}
			// Parent must be the bot key
			parent := got.Parent()
			if parent == nil {
				t.Fatal("NewBotChatKey() should have a bot parent key")
			}
			if parent.Collection() != botsCollection {
				t.Errorf("NewBotChatKey() parent collection = %q, want %q", parent.Collection(), botsCollection)
			}
			if parent.ID != tt.botID {
				t.Errorf("NewBotChatKey() parent ID = %v, want %v", parent.ID, tt.botID)
			}
			// Grandparent must be the platform key
			grandParent := parent.Parent()
			if grandParent == nil {
				t.Fatal("NewBotChatKey() should have a platform grandparent key")
			}
			if grandParent.Collection() != botPlatformsCollection {
				t.Errorf("NewBotChatKey() grandparent collection = %q, want %q", grandParent.Collection(), botPlatformsCollection)
			}
		})
	}
}

// testChatData is a minimal in-memory implementation of BotChatData for testing.
type testChatData struct {
	botsfwmodels.ChatBaseData
}

func newTestChatData() botsfwmodels.BotChatData {
	return &testChatData{}
}

func TestGetBotChat(t *testing.T) {
	ctx := context.Background()
	const platform = botsfwconst.Platform("telegram")
	const botID = "bot1"
	const chatID = "chat99"

	t.Run("not_found", func(t *testing.T) {
		db := dalgo2memory.NewDB()
		var chat interface{}
		err := db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) error {
			var innerErr error
			_, innerErr = GetBotChat(ctx, tx, platform, botID, chatID, newTestChatData)
			return innerErr
		})
		_ = chat
		if err == nil {
			t.Error("GetBotChat() expected not-found error, got nil")
		}
		if !dal.IsNotFound(err) {
			t.Errorf("GetBotChat() expected not-found error, got: %v", err)
		}
	})

	t.Run("found_after_insert", func(t *testing.T) {
		db := dalgo2memory.NewDB()

		// Insert a chat record
		key := NewBotChatKey(platform, botID, chatID)
		data := &testChatData{}
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			rec := dal.NewRecordWithData(key, data)
			return tx.Insert(ctx, rec)
		})
		if err != nil {
			t.Fatalf("failed to insert chat record: %v", err)
		}

		// Now read it back
		err = db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) error {
			result, innerErr := GetBotChat(ctx, tx, platform, botID, chatID, newTestChatData)
			if innerErr != nil {
				return innerErr
			}
			if result.ID != chatID {
				t.Errorf("GetBotChat() ID = %v, want %v", result.ID, chatID)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetBotChat() unexpected error: %v", err)
		}
	})
}
