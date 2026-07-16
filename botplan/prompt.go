package botplan

import (
	"fmt"

	"github.com/bots-go-framework/bots-fw/botstoken"
)

// MaxChoiceTokenBytes is the maximum byte length of a Choice.Token.
//
// It matches botstoken.MaxTokenBytes (64), which is Telegram's callback_data
// ceiling (capability-map telegram/callback-query constraints.callbackDataBytes
// = 64). A token that fits 64 bytes also fits WhatsApp's 256-char interactive id
// (capability-map whatsapp/reply-buttons semanticsPayloadBytes = 256) and a
// wa.me "?text=" suffix, so 64 is the binding constraint for all platforms.
const MaxChoiceTokenBytes = botstoken.MaxTokenBytes

// Choice is one selectable option in an ActionPrompt.
//
// Label is what the user sees. Token is the botstoken-encoded callback state
// (verb/subject/args) that identifies the choice when tapped; it must be ≤64
// bytes. A choice with an empty Token is rejected by Validate — a prompt whose
// taps carry no state is a bug, not a degradation.
type Choice struct {
	Label string
	Token string
}

// ActionPrompt presents N choices and expects one back, each carrying ≤64 bytes
// of callback state (capability vocabulary: prompt-choices).
//
// LayoutRows is a hint for how many rows to arrange the choices into on
// platforms that support a button grid (Telegram). It is advisory: WhatsApp has
// no grid, so its renderer ignores it (capability-map whatsapp/reply-buttons
// grid=false). Zero or negative means "let the renderer decide" (one button per
// row on Telegram).
type ActionPrompt struct {
	Choices    []Choice
	LayoutRows int
}

// Validate checks the prompt's invariants: at least one choice, every label
// non-empty, every token non-empty and ≤64 bytes.
//
// It does NOT decode the token — a token is an opaque botstoken string here; the
// only structural rule the neutral layer enforces is the byte ceiling, because
// that is the cross-platform fact (see MaxChoiceTokenBytes).
func (p ActionPrompt) Validate() error {
	if len(p.Choices) == 0 {
		return fmt.Errorf("%w: prompt has no choices", ErrInvalidPlan)
	}
	for i, c := range p.Choices {
		if c.Label == "" {
			return fmt.Errorf("%w: choice %d has an empty label", ErrInvalidPlan, i)
		}
		if c.Token == "" {
			return fmt.Errorf("%w: choice %d (%q) has an empty token", ErrInvalidPlan, i, c.Label)
		}
		if n := len(c.Token); n > MaxChoiceTokenBytes {
			return fmt.Errorf("%w: choice %d (%q) token is %d bytes, max %d",
				ErrTokenTooLong, i, c.Label, n, MaxChoiceTokenBytes)
		}
	}
	return nil
}

// URLAction is one labelled tappable web link carrying a token in its URL
// (capability vocabulary: url-action).
//
// On Telegram it is a URL inline button; on WhatsApp in-window it is an
// interactive cta_url button (capability-map whatsapp/cta-url-button), and out
// of window it can only ride a template's approved URL button
// (capability-map whatsapp/template-buttons urlButtonBaseFixedAtApproval).
type URLAction struct {
	Label string
	URL   string
}

// Validate checks a URLAction has both a label and a URL.
func (u URLAction) Validate() error {
	if u.Label == "" {
		return fmt.Errorf("%w: URL action has an empty label", ErrInvalidPlan)
	}
	if u.URL == "" {
		return fmt.Errorf("%w: URL action %q has an empty URL", ErrInvalidPlan, u.Label)
	}
	return nil
}
