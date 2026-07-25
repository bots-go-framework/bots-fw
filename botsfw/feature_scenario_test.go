package botsfw

import (
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

func TestFeatureScenarioHarnessDedicatedAndEmbedded(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		mode       FeatureMountMode
		navigation []NavigatorID
	}{
		{"dedicated", FeatureMountDedicated, []NavigatorID{"debtus"}},
		{"embedded", FeatureMountEmbedded, []NavigatorID{"home", "debtus"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			mount := FeatureMount{ID: "debtus", Mode: scenario.mode}
			if scenario.mode == FeatureMountEmbedded {
				mount.Namespace = "debtus"
			}
			err := AssertFeatureScenario(FeatureScenario{Mount: mount, Messages: []botmsg.MessageFromBot{{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}}, Navigation: scenario.navigation, SideEffects: []string{"show-space"}}, FeatureScenarioExpectation{Mode: scenario.mode, Messages: 1, Navigation: scenario.navigation, SideEffects: []string{"show-space"}, Policy: PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardHostOnly, HostMayShowBottomKeyboard: true}})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFeatureScenarioHarnessRejectsKeyboardPolicyViolation(t *testing.T) {
	err := AssertFeatureScenario(FeatureScenario{Mount: FeatureMount{ID: "d", Mode: FeatureMountDedicated}, Messages: []botmsg.MessageFromBot{{Presentation: botmsg.Presentation{PersistentBottomKeyboard: true}}}}, FeatureScenarioExpectation{Mode: FeatureMountDedicated, Messages: 1, Policy: PresentationPolicy{PersistentBottomKeyboard: PersistentBottomKeyboardDeny}})
	if err == nil || !strings.Contains(err.Error(), "persistent") {
		t.Fatalf("expected policy violation, got %v", err)
	}
}
