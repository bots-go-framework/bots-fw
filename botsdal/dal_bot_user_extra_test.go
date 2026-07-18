package botsdal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
)

func TestNewPlatformUserKey(t *testing.T) {
	tests := []struct {
		name        string
		platformID  botsfwconst.Platform
		botUserID   string
		shouldPanic bool
	}{
		{
			name:       "valid",
			platformID: "telegram",
			botUserID:  "user123",
		},
		{
			name:        "empty_botUserID_panics",
			platformID:  "telegram",
			botUserID:   "",
			shouldPanic: true,
		},
		{
			name:        "empty_platform_panics",
			platformID:  "",
			botUserID:   "user123",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewPlatformUserKey() did not panic")
					}
				}()
			}
			got := NewPlatformUserKey(tt.platformID, tt.botUserID)
			if tt.shouldPanic {
				t.Errorf("NewPlatformUserKey() should have panicked")
				return
			}
			if got == nil {
				t.Fatal("NewPlatformUserKey() returned nil")
			}
			if got.Collection() != botUsersCollection {
				t.Errorf("NewPlatformUserKey() collection = %q, want %q", got.Collection(), botUsersCollection)
			}
			if got.ID != tt.botUserID {
				t.Errorf("NewPlatformUserKey() ID = %v, want %v", got.ID, tt.botUserID)
			}
			parent := got.Parent()
			if parent == nil {
				t.Fatal("NewPlatformUserKey() should have a platform parent key")
			}
			if parent.Collection() != botPlatformsCollection {
				t.Errorf("NewPlatformUserKey() parent collection = %q, want %q", parent.Collection(), botPlatformsCollection)
			}
		})
	}
}

// simplePlatformUser is a minimal implementation of PlatformUserData that does NOT embed
// PlatformUserBaseDbo (which has strict Validate() requirements) so tests can
// exercise CreatePlatformUserRecord without satisfying all validation rules.
type simplePlatformUser struct{}

func (u *simplePlatformUser) BaseData() *botsfwmodels.PlatformUserBaseDbo { return nil }
func (u *simplePlatformUser) GetAppUserID() string                        { return "" }
func (u *simplePlatformUser) SetAppUserID(_ string)                       {}
func (u *simplePlatformUser) IsAccessGranted() bool                       { return false }
func (u *simplePlatformUser) SetAccessGranted(_ bool) bool                { return false }
func (u *simplePlatformUser) SetUpdatedTime(_ time.Time)                  {}

var _ botsfwmodels.PlatformUserData = (*simplePlatformUser)(nil)

func TestGetPlatformUser_WithMemoryDB(t *testing.T) {
	ctx := context.Background()
	const platform = botsfwconst.Platform("telegram")
	const botUserID = "user42"

	t.Run("not_found", func(t *testing.T) {
		db := dalgo2memory.NewDB()
		err := db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) error {
			userData := &simplePlatformUser{}
			_, innerErr := GetPlatformUser(ctx, tx, platform, botUserID, userData)
			return innerErr
		})
		if !dal.IsNotFound(err) {
			t.Errorf("GetPlatformUser() expected not-found error, got: %v", err)
		}
	})

	t.Run("found_after_insert", func(t *testing.T) {
		db := dalgo2memory.NewDB()

		// Insert a user record directly using the DAL key
		key := NewPlatformUserKey(platform, botUserID)
		insertData := &simplePlatformUser{}
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			rec := dal.NewRecordWithData(key, insertData)
			return tx.Insert(ctx, rec)
		})
		if err != nil {
			t.Fatalf("failed to insert user record: %v", err)
		}

		// Now read it back via GetPlatformUser
		err = db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) error {
			readData := &simplePlatformUser{}
			result, innerErr := GetPlatformUser(ctx, tx, platform, botUserID, readData)
			if innerErr != nil {
				return innerErr
			}
			if result.ID != botUserID {
				t.Errorf("GetPlatformUser() ID = %v, want %v", result.ID, botUserID)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("GetPlatformUser() unexpected error: %v", err)
		}
	})
}

func TestCreatePlatformUserRecord_WithMemoryDB(t *testing.T) {
	ctx := context.Background()
	const platform = botsfwconst.Platform("telegram")

	t.Run("creates_successfully", func(t *testing.T) {
		db := dalgo2memory.NewDB()
		const botUserID = "newuser1"
		// simplePlatformUser has no Validate() method, so the validation branch is skipped
		userData := &simplePlatformUser{}

		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return CreatePlatformUserRecord(ctx, tx, platform, botUserID, userData)
		})
		if err != nil {
			t.Fatalf("CreatePlatformUserRecord() unexpected error: %v", err)
		}

		// Verify we can read it back
		err = db.RunReadonlyTransaction(ctx, func(ctx context.Context, tx dal.ReadTransaction) error {
			readData := &simplePlatformUser{}
			result, innerErr := GetPlatformUser(ctx, tx, platform, botUserID, readData)
			if innerErr != nil {
				return innerErr
			}
			if result.ID != botUserID {
				t.Errorf("after CreatePlatformUserRecord(), ID = %v, want %v", result.ID, botUserID)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("reading created user failed: %v", err)
		}
	})

	// The panic for empty botUserID occurs synchronously inside the tx worker,
	// before RunReadwriteTransaction returns, so we can recover it directly.
	t.Run("empty_botUserID_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("CreatePlatformUserRecord() did not panic with empty botUserID")
			}
		}()
		userData := &simplePlatformUser{}
		// The panic occurs before any DB interaction (nil tx is fine since panic is first)
		_ = CreatePlatformUserRecord(context.Background(), nil, platform, "", userData)
	})
}

// validatingPlatformUser is a PlatformUserData whose Validate() returns an error.
type validatingPlatformUser struct {
	simplePlatformUser
	validateErr error
}

func (v *validatingPlatformUser) Validate() error {
	return v.validateErr
}

func TestCreatePlatformUserRecord_ValidationError(t *testing.T) {
	ctx := context.Background()
	const platform = botsfwconst.Platform("telegram")
	const botUserID = "user99"

	db := dalgo2memory.NewDB()

	wantErr := errors.New("validation failed")
	userData := &validatingPlatformUser{validateErr: wantErr}

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return CreatePlatformUserRecord(ctx, tx, platform, botUserID, userData)
	})
	if err == nil {
		t.Error("CreatePlatformUserRecord() expected validation error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("CreatePlatformUserRecord() error = %v, want %v", err, wantErr)
	}
}
