package botsfw

import (
	"fmt"

	"github.com/bots-go-framework/bots-fw/botmsg"
)

// ScenarioMessagePath identifies the transport path used to present a
// message. Feature-returned messages and direct sends are feature-owned;
// HostSend is available only to the host fixture for its own UI.
type ScenarioMessagePath string

const (
	ScenarioRouterReturn ScenarioMessagePath = "router-return"
	ScenarioDirectSend   ScenarioMessagePath = "direct-send"
	ScenarioHostSend     ScenarioMessagePath = "host-send"
)

type ScenarioMessage struct {
	Path    ScenarioMessagePath
	Message botmsg.MessageFromBot
}

// FeatureScenario is an executed, deterministic host-level acceptance trace.
// It records both router-return and direct-send paths so presentation policy
// cannot be bypassed by choosing a different response path.
type FeatureScenario struct {
	Mount       FeatureMount
	Messages    []ScenarioMessage
	Navigation  []NavigatorID
	SideEffects []string
}

// FeatureScenarioFlow is supplied by a feature adapter fixture. The harness
// executes it under a concrete mount and validates every emitted message.
type FeatureScenarioFlow func(*FeatureScenarioContext) error

// FeatureScenarioContext provides the only output paths a replay fixture may
// use. SendHost models host UI (Home, keyboard removal, or inline controls),
// while feature paths are always treated as non-host-owned.
type FeatureScenarioContext struct {
	scenario *FeatureScenario
	policy   PresentationPolicy
}

func (c *FeatureScenarioContext) ReturnFromRouter(message botmsg.MessageFromBot) error {
	return c.addMessage(ScenarioRouterReturn, message, false)
}

func (c *FeatureScenarioContext) SendDirect(message botmsg.MessageFromBot) error {
	return c.addMessage(ScenarioDirectSend, message, false)
}

func (c *FeatureScenarioContext) SendHost(message botmsg.MessageFromBot) error {
	return c.addMessage(ScenarioHostSend, message, true)
}

func (c *FeatureScenarioContext) Navigate(to NavigatorID) {
	c.scenario.Navigation = append(c.scenario.Navigation, to)
}

func (c *FeatureScenarioContext) SideEffect(effect string) {
	c.scenario.SideEffects = append(c.scenario.SideEffects, effect)
}

func (c *FeatureScenarioContext) addMessage(path ScenarioMessagePath, message botmsg.MessageFromBot, hostOwned bool) error {
	if err := c.policy.Validate(message, hostOwned); err != nil {
		return fmt.Errorf("%s message violates presentation policy: %w", path, err)
	}
	c.scenario.Messages = append(c.scenario.Messages, ScenarioMessage{Path: path, Message: message})
	return nil
}

// ReplayFeatureScenario validates a mount and executes the supplied flow under
// that mount context. It is intentionally transport-free, but exercises the
// same router-return and direct-send ownership boundaries as a host adapter.
func ReplayFeatureScenario(mount FeatureMount, policy PresentationPolicy, flow FeatureScenarioFlow) (FeatureScenario, error) {
	if _, err := NewFeatureRegistry(mount); err != nil {
		return FeatureScenario{}, err
	}
	scenario := FeatureScenario{Mount: mount}
	if flow == nil {
		return scenario, nil
	}
	if err := flow(&FeatureScenarioContext{scenario: &scenario, policy: policy}); err != nil {
		return FeatureScenario{}, err
	}
	return scenario, nil
}

type FeatureScenarioExpectation struct {
	Mode        FeatureMountMode
	Messages    int
	Paths       []ScenarioMessagePath
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
	if len(want.Paths) > 0 {
		if len(s.Messages) != len(want.Paths) {
			return fmt.Errorf("message paths = %d, want %d", len(s.Messages), len(want.Paths))
		}
		for i, message := range s.Messages {
			if message.Path != want.Paths[i] {
				return fmt.Errorf("message %d path = %q, want %q", i, message.Path, want.Paths[i])
			}
		}
	}
	if !sameNavigators(s.Navigation, want.Navigation) {
		return fmt.Errorf("navigation = %v, want %v", s.Navigation, want.Navigation)
	}
	if !sameStrings(s.SideEffects, want.SideEffects) {
		return fmt.Errorf("side effects = %v, want %v", s.SideEffects, want.SideEffects)
	}
	for _, message := range s.Messages {
		if err := want.Policy.Validate(message.Message, message.Path == ScenarioHostSend); err != nil {
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
