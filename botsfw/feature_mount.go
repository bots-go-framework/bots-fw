package botsfw

import (
	"fmt"
	"strings"

	"github.com/bots-go-framework/bots-fw/botinput"
)

// FeatureMountMode describes how a feature is exposed by its host bot.
// Dedicated features own the bot surface; embedded features share it with the
// host and therefore must use an unambiguous namespace and command ownership.
type FeatureMountMode string

type CapabilityID string
type NavigatorID string

const (
	FeatureMountDedicated FeatureMountMode = "dedicated"
	FeatureMountEmbedded  FeatureMountMode = "embedded"
)

// CommandOwnership declares a command code claimed by a feature for one or
// more input types. It is metadata only: registration remains compatible with
// existing Router.RegisterCommands callers.
type CommandOwnership struct {
	Code               CommandCode
	Aliases            []CommandCode
	CallbackNamespaces []string
	InputTypes         []botinput.Type
}

// FeatureMount is the stable host-to-feature integration contract. Navigator
// and Capabilities are intentionally strings so individual bot applications can
// evolve independently while hosts can validate and discover their surfaces.
type FeatureMount struct {
	ID           string
	Mode         FeatureMountMode
	Namespace    string
	Navigator    NavigatorID
	Capabilities []CapabilityID
	Commands     []CommandOwnership
}

const (
	ReservedCancelCommand CommandCode = "cancel"
	ReservedHomeCommand   CommandCode = "home"
)

// FeatureRegistry is the host startup registry. Its constructor validates
// mounted features before a router can be exposed, while old hosts may adopt it
// without changing BotProfile or Router APIs.
type FeatureRegistry struct{ mounts []FeatureMount }

func NewFeatureRegistry(mounts ...FeatureMount) (*FeatureRegistry, error) {
	if err := ValidateFeatureMounts(mounts); err != nil {
		return nil, err
	}
	return &FeatureRegistry{mounts: append([]FeatureMount(nil), mounts...)}, nil
}

func (r *FeatureRegistry) Mounts() []FeatureMount { return append([]FeatureMount(nil), r.mounts...) }

// ValidateFeatureMounts fails before startup if mounted features would make
// routing ambiguous. This is deliberately separate from Router so existing
// bots can adopt feature descriptors incrementally.
func ValidateFeatureMounts(mounts []FeatureMount) error {
	ids := make(map[string]struct{}, len(mounts))
	namespaces := make(map[string]struct{}, len(mounts))
	commands := make(map[featureCommandKey]string)
	callbackNamespaces := make(map[string]string)
	for _, mount := range mounts {
		if err := validateFeatureMount(mount); err != nil {
			return err
		}
		if _, exists := ids[mount.ID]; exists {
			return fmt.Errorf("duplicate feature mount ID %q", mount.ID)
		}
		ids[mount.ID] = struct{}{}
		if mount.Namespace != "" {
			if _, exists := namespaces[mount.Namespace]; exists {
				return fmt.Errorf("duplicate feature mount namespace %q", mount.Namespace)
			}
			namespaces[mount.Namespace] = struct{}{}
		}
		for _, ownership := range mount.Commands {
			for _, code := range append([]CommandCode{ownership.Code}, ownership.Aliases...) {
				if err := registerFeatureCommand(commands, mount.ID, ownership.InputTypes, code); err != nil {
					return err
				}
			}
			for _, namespace := range ownership.CallbackNamespaces {
				if owner, exists := callbackNamespaces[namespace]; exists {
					return fmt.Errorf("callback namespace %q is owned by both features %q and %q", namespace, owner, mount.ID)
				}
				callbackNamespaces[namespace] = mount.ID
			}
		}
	}
	return nil
}

func registerFeatureCommand(commands map[featureCommandKey]string, mountID string, inputTypes []botinput.Type, code CommandCode) error {
	if code == ReservedCancelCommand || code == ReservedHomeCommand {
		return fmt.Errorf("feature %q cannot claim reserved host command %q", mountID, code)
	}
	for _, inputType := range inputTypes {
		key := featureCommandKey{inputType: inputType, code: code}
		if owner, exists := commands[key]; exists {
			return fmt.Errorf("command %q for input type %q is owned by both features %q and %q", code, inputType, owner, mountID)
		}
		commands[key] = mountID
	}
	return nil
}

type featureCommandKey struct {
	inputType botinput.Type
	code      CommandCode
}

func validateFeatureMount(mount FeatureMount) error {
	if strings.TrimSpace(mount.ID) == "" {
		return fmt.Errorf("feature mount ID is required")
	}
	switch mount.Mode {
	case FeatureMountDedicated, FeatureMountEmbedded:
	default:
		return fmt.Errorf("feature %q has invalid mount mode %q", mount.ID, mount.Mode)
	}
	if mount.Mode == FeatureMountEmbedded && strings.TrimSpace(mount.Namespace) == "" {
		return fmt.Errorf("embedded feature %q requires a namespace", mount.ID)
	}
	for _, ownership := range mount.Commands {
		if ownership.Code == "" {
			return fmt.Errorf("feature %q has command ownership with an empty code", mount.ID)
		}
		if len(ownership.InputTypes) == 0 {
			return fmt.Errorf("feature %q command %q has no input types", mount.ID, ownership.Code)
		}
		for _, alias := range ownership.Aliases {
			if alias == "" {
				return fmt.Errorf("feature %q command %q has an empty alias", mount.ID, ownership.Code)
			}
		}
		for _, namespace := range ownership.CallbackNamespaces {
			if strings.TrimSpace(namespace) == "" {
				return fmt.Errorf("feature %q command %q has an empty callback namespace", mount.ID, ownership.Code)
			}
		}
	}
	return nil
}
