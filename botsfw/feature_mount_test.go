package botsfw

import (
	"strings"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
)

func TestValidateFeatureMounts(t *testing.T) {
	valid := []FeatureMount{
		{ID: "home", Mode: FeatureMountDedicated, Commands: []CommandOwnership{{Code: "dashboard", InputTypes: []botinput.Type{botinput.TypeText}}}},
		{ID: "debtus", Mode: FeatureMountEmbedded, Namespace: "debtus", Navigator: "space", Capabilities: []CapabilityID{"receipts"}, Commands: []CommandOwnership{{Code: "debtus_receipt", InputTypes: []botinput.Type{botinput.TypeCallbackQuery}}}},
	}
	if err := ValidateFeatureMounts(valid); err != nil {
		t.Fatalf("valid mounts: %v", err)
	}

	for name, mounts := range map[string][]FeatureMount{
		"duplicate ID": {
			valid[0], {ID: "home", Mode: FeatureMountDedicated},
		},
		"embedded namespace required": {
			{ID: "debtus", Mode: FeatureMountEmbedded},
		},
		"duplicate command": {
			{ID: "one", Mode: FeatureMountDedicated, Commands: []CommandOwnership{{Code: "open", InputTypes: []botinput.Type{botinput.TypeText}}}},
			{ID: "two", Mode: FeatureMountEmbedded, Namespace: "two", Commands: []CommandOwnership{{Code: "open", InputTypes: []botinput.Type{botinput.TypeText}}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateFeatureMounts(mounts); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateFeatureMountsErrorsNameTheConflict(t *testing.T) {
	err := ValidateFeatureMounts([]FeatureMount{
		{ID: "one", Mode: FeatureMountEmbedded, Namespace: "shared"},
		{ID: "two", Mode: FeatureMountEmbedded, Namespace: "shared"},
	})
	if err == nil || !strings.Contains(err.Error(), "shared") {
		t.Fatalf("expected namespace conflict, got %v", err)
	}
}
