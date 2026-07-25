package botsfw

import (
	"fmt"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

// FeatureScenario is a deterministic host-level acceptance harness. Feature
// adapters provide the plan instead of invoking network responders, so the
// same assertions cover dedicated and embedded mounting modes.
type FeatureScenario struct {
	Mount       FeatureMount
	Messages    []botmsg.MessageFromBot
	Navigation  []NavigatorID
	SideEffects []string
}

type FeatureScenarioExpectation struct {
	Mode        FeatureMountMode
	Messages    int
	Navigation  []NavigatorID
	SideEffects []string
	Policy      PresentationPolicy
}

func AssertFeatureScenario(s FeatureScenario, want FeatureScenarioExpectation) error {
	if s.Mount.Mode != want.Mode {
		return fmt.Errorf("mount mode = %q, want %q", s.Mount.Mode, want.Mode)
	}
	if len(s.Messages) != want.Messages {
		return fmt.Errorf("messages = %d, want %d", len(s.Messages), want.Messages)
	}
	if !sameNavigators(s.Navigation, want.Navigation) {
		return fmt.Errorf("navigation = %v, want %v", s.Navigation, want.Navigation)
	}
	if !sameStrings(s.SideEffects, want.SideEffects) {
		return fmt.Errorf("side effects = %v, want %v", s.SideEffects, want.SideEffects)
	}
	for _, message := range s.Messages {
		if err := want.Policy.Validate(message); err != nil {
			return fmt.Errorf("message plan violates presentation policy: %w", err)
		}
	}
	return nil
}

func sameNavigators(a, b []NavigatorID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
