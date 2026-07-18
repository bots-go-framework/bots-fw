package botswebhook

import (
	"context"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"go.uber.org/mock/gomock"
)

// logInputReferralNoText is a botinput.ReferralMessage that does NOT implement
// botinput.TextMessage (a real referral has no Text()) — so it actually reaches
// logInput's `case botinput.ReferralMessage`, unlike the TextMessage-implementing
// helper in coverage_extra_test.go. Reuses logInputBase from that file.
type logInputReferralNoText struct {
	*logInputBase
}

func (l *logInputReferralNoText) Type() string    { return "ad" }
func (l *logInputReferralNoText) Source() string  { return "SHORTLINK" }
func (l *logInputReferralNoText) RefData() string { return "promo123" }

// TestLogInput_ReferralMessage_DoesNotPanic is the regression test for the
// logInput bug: the ReferralMessage case did `input.(botinput.TextMessage).Text()`,
// which panics for a referral that has no Text(). This test panics (fails) on the
// old code and passes with the fix (which logs Type/Source/RefData).
func TestLogInput_ReferralMessage_DoesNotPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	ref := &logInputReferralNoText{logInputBase: newLogInputBase(ctrl, botinput.TypeReferral)}
	// A panic here fails the test (Go reports the panic as a test failure).
	webhookDriver{}.logInput(context.Background(), 0, ref)
}

// TestRegisterCommands_LocationActionWithoutTypeLocation is the regression test
// for the RegisterCommands bug: the post-loop location add used `&& locationAdded`
// instead of `&& !locationAdded` (as every sibling does), so a command carrying a
// LocationAction but not listing TypeLocation in InputTypes was never registered
// to handle location messages. Fails on the old code (no TypeLocation command),
// passes with the fix.
func TestRegisterCommands_LocationActionWithoutTypeLocation(t *testing.T) {
	router := NewWebhookRouter(nil).(*webhooksRouter)
	router.RegisterCommands(botsfw.Command{
		Code:           "share_loc",
		InputTypes:     []botinput.Type{botinput.TypeText},
		TextAction:     dummyTextAction,
		LocationAction: dummyLocationAction,
	})
	cmds := router.RegisteredCommands()
	if len(cmds[botinput.TypeLocation]) == 0 {
		t.Fatal("a command with a LocationAction must be registered for TypeLocation even when TypeLocation is not in InputTypes")
	}
	if _, ok := cmds[botinput.TypeLocation]["share_loc"]; !ok {
		t.Fatalf("expected 'share_loc' registered for TypeLocation, got %+v", cmds[botinput.TypeLocation])
	}
}
