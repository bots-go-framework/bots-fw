package botplan

import (
	"errors"
	"testing"
)

func TestMessagePlanValidate(t *testing.T) {
	validPrompt := &ActionPrompt{Choices: []Choice{{Label: "Yes", Token: "y"}}}

	tests := []struct {
		name    string
		plan    MessagePlan
		wantErr error
	}{
		{
			name:    "empty plan (no text, no media)",
			plan:    MessagePlan{},
			wantErr: ErrInvalidPlan,
		},
		{
			name:    "text only is valid",
			plan:    MessagePlan{Text: RichText("hi")},
			wantErr: nil,
		},
		{
			name:    "media only (no text) is valid",
			plan:    MessagePlan{Media: &MediaRef{ImageURL: "https://x.io/a.jpg"}},
			wantErr: nil,
		},
		{
			name:    "text + valid prompt",
			plan:    MessagePlan{Text: RichText("hi"), Prompt: validPrompt},
			wantErr: nil,
		},
		{
			name: "text + invalid prompt propagates",
			plan: MessagePlan{Text: RichText("hi"), Prompt: &ActionPrompt{
				Choices: []Choice{{Label: "", Token: "t"}},
			}},
			wantErr: ErrInvalidPlan,
		},
		{
			name: "prompt + urlaction coexist (renderer resolves)",
			plan: MessagePlan{
				Text:      RichText("hi"),
				Prompt:    validPrompt,
				URLAction: &URLAction{Label: "View", URL: "https://x.io"},
			},
			wantErr: nil,
		},
		{
			name: "livepanel + proactive is incoherent",
			plan: MessagePlan{
				Text:      RichText("hi"),
				LivePanel: &LivePanel{PanelKey: "card1"},
				Proactive: &ProactiveSpec{Purpose: "notice"},
			},
			wantErr: ErrInvalidPlan,
		},
		{
			name: "livepanel + prompt (edit a panel with buttons) is valid",
			plan: MessagePlan{
				Text:      RichText("hi"),
				LivePanel: &LivePanel{PanelKey: "card1"},
				Prompt:    validPrompt,
			},
			wantErr: nil,
		},
		{
			name: "invalid urlaction propagates",
			plan: MessagePlan{
				Text:      RichText("hi"),
				URLAction: &URLAction{Label: "View"},
			},
			wantErr: ErrInvalidPlan,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want errors.Is %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessagePlanIsProactive(t *testing.T) {
	if (MessagePlan{}).IsProactive() {
		t.Error("nil Proactive should be a reply")
	}
	if !(MessagePlan{Proactive: &ProactiveSpec{}}).IsProactive() {
		t.Error("non-nil Proactive should be proactive")
	}
}
