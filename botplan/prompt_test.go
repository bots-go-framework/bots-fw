package botplan

import (
	"errors"
	"strings"
	"testing"
)

func TestActionPromptValidate(t *testing.T) {
	// A 64-byte token (exactly at the ceiling) is valid; 65 is not.
	tokenAt64 := strings.Repeat("a", MaxChoiceTokenBytes)
	tokenOver := strings.Repeat("a", MaxChoiceTokenBytes+1)

	tests := []struct {
		name    string
		prompt  ActionPrompt
		wantErr error
	}{
		{
			name:    "no choices",
			prompt:  ActionPrompt{},
			wantErr: ErrInvalidPlan,
		},
		{
			name:    "empty label",
			prompt:  ActionPrompt{Choices: []Choice{{Label: "", Token: "t"}}},
			wantErr: ErrInvalidPlan,
		},
		{
			name:    "empty token",
			prompt:  ActionPrompt{Choices: []Choice{{Label: "Yes", Token: ""}}},
			wantErr: ErrInvalidPlan,
		},
		{
			name:    "token exactly 64 bytes ok",
			prompt:  ActionPrompt{Choices: []Choice{{Label: "Yes", Token: tokenAt64}}},
			wantErr: nil,
		},
		{
			name:    "token 65 bytes too long",
			prompt:  ActionPrompt{Choices: []Choice{{Label: "Yes", Token: tokenOver}}},
			wantErr: ErrTokenTooLong,
		},
		{
			name: "valid multi-choice",
			prompt: ActionPrompt{Choices: []Choice{
				{Label: "I'll be there", Token: "rsvp_i_1"},
				{Label: "Maybe", Token: "rsvp_m_1"},
				{Label: "Can't make it", Token: "rsvp_n_1"},
			}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prompt.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want errors.Is %v", err, tt.wantErr)
			}
			// ErrTokenTooLong must also satisfy ErrInvalidPlan for callers matching the base.
			if tt.wantErr == ErrTokenTooLong && !errors.Is(err, ErrInvalidPlan) {
				t.Errorf("token-too-long error should wrap ErrInvalidPlan")
			}
		})
	}
}

func TestURLActionValidate(t *testing.T) {
	tests := []struct {
		name    string
		action  URLAction
		wantErr bool
	}{
		{"valid", URLAction{Label: "View", URL: "https://x.io"}, false},
		{"empty label", URLAction{URL: "https://x.io"}, true},
		{"empty url", URLAction{Label: "View"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr && !errors.Is(err, ErrInvalidPlan) {
				t.Errorf("want ErrInvalidPlan, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("want nil, got %v", err)
			}
		})
	}
}

func TestMaxChoiceTokenBytesMatchesTelegramCallbackDataLimit(t *testing.T) {
	// Documents the load-bearing fact: 64 is Telegram's callback_data ceiling
	// (capability-map telegram/callback-query constraints.callbackDataBytes).
	if MaxChoiceTokenBytes != 64 {
		t.Errorf("MaxChoiceTokenBytes = %d, want 64", MaxChoiceTokenBytes)
	}
}
