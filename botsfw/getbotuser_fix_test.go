package botsfw

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
)

// TestGetBotUserForUpdate_UsesCallerTxNoNestedDeadlock is the regression test for
// the GetBotUserForUpdate bug. The old code ignored its tx parameter and opened a
// NESTED whcb.db.RunReadwriteTransaction; calling it from within a caller's
// transaction (which the tx parameter implies) then deadlocks (an in-memory DB) or
// is unsupported (backends that forbid nested transactions), and the "for update"
// read never joins the caller's tx.
//
// platformUser.Data is pre-seeded so getBotUser returns without a DB read — the
// ONLY behaviour under test is that GetBotUserForUpdate does not open a nested
// transaction. Old code: nested tx under the open outer tx → deadlock/error.
// Fixed code: uses the passed tx → returns the user immediately.
func TestGetBotUserForUpdate_UsesCallerTxNoNestedDeadlock(t *testing.T) {
	whcb := newCoverageWHCB(t)
	whcb.db = dalgo2memory.NewDB()
	whcb.platformUser.Data = &simplePlatformUser{appUserID: "u1"}

	ctx := context.Background()
	var got botsdal.BotUser
	err := whcb.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		var e error
		got, e = whcb.GetBotUserForUpdate(ctx, tx)
		return e
	})
	if err != nil {
		t.Fatalf("GetBotUserForUpdate within a caller's tx must not deadlock/error: %v", err)
	}
	if got.Data == nil || got.Data.GetAppUserID() != "u1" {
		t.Fatalf("expected the cached bot user (u1), got %+v", got.Data)
	}
}
