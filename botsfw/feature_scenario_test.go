package botsfw

import (
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-go-core/botkb"
)

func TestFeatureScenarioReplayDedicatedAndEmbedded(t *testing.T) {
	debtusFlow := func(ctx *FeatureScenarioContext) error {
		if err := ctx.ReturnFromRouter(botmsg.MessageFromBot{}); err != nil {
			return err
		}
		if err := ctx.SendDirect(botmsg.MessageFromBot{}); err != nil {
			return err
		}
		ctx.SideEffect("show-space")
		return nil
	}
	for _, test := range []struct {
		name       string
		mount      FeatureMount
		navigation []NavigatorID
	}{
		{"dedicated", FeatureMount{ID: "debtus", Mode: FeatureMountDedicated}, []NavigatorID{"debtus"}},
		{"embedded", FeatureMount{ID: "debtus-embedded", Mode: FeatureMountEmbedded, Namespace: "debtus"}, []NavigatorID{"home", "debtus"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			scenario, err := ReplayFeatureScenario(test.mount, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}, func(ctx *FeatureScenarioContext) error {
				ctx.scenario.Navigation = append(ctx.scenario.Navigation, test.navigation...)
				return debtusFlow(ctx)
			})
			if err != nil {
				t.Fatal(err)
			}
			err = AssertFeatureScenario(scenario, FeatureScenarioExpectation{
				Mode:        test.mount.Mode,
				Messages:    2,
				Paths:       []ScenarioMessagePath{ScenarioRouterReturn, ScenarioDirectSend},
				Navigation:  test.navigation,
				SideEffects: []string{"show-space"},
				Policy:      PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly},
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFeatureScenarioReplayRejectsEmbeddedFeatureBottomKeyboard(t *testing.T) {
	_, err := ReplayFeatureScenario(
		FeatureMount{ID: "debtus-embedded", Mode: FeatureMountEmbedded, Namespace: "debtus"},
		PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly},
		func(ctx *FeatureScenarioContext) error {
			message := botmsg.MessageFromBot{}
			message.Keyboard = botkb.NewMessageKeyboard(botkb.KeyboardTypeBottom)
			return ctx.ReturnFromRouter(message)
		},
	)
	if err == nil || !strings.Contains(err.Error(), "reserved for the host") {
		t.Fatalf("expected embedded feature keyboard rejection, got %v", err)
	}
}

func TestFeatureScenarioReplayAllowsHostHomeRemoveAndInlineMessages(t *testing.T) {
	scenario, err := ReplayFeatureScenario(FeatureMount{ID: "debtus", Mode: FeatureMountDedicated}, PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly}, func(ctx *FeatureScenarioContext) error {
		for _, kind := range []botkb.KeyboardType{botkb.KeyboardTypeBottom, botkb.KeyboardTypeHide, botkb.KeyboardTypeInline} {
			message := botmsg.MessageFromBot{}
			message.Keyboard = botkb.NewMessageKeyboard(kind)
			if err := ctx.SendHost(message); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := AssertFeatureScenario(scenario, FeatureScenarioExpectation{
		Mode:     FeatureMountDedicated,
		Messages: 3,
		Paths:    []ScenarioMessagePath{ScenarioHostSend, ScenarioHostSend, ScenarioHostSend},
		Policy:   PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly},
	}); err != nil {
		t.Fatal(err)
	}
}
