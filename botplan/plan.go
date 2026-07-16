package botplan

import "fmt"

// MessagePlan is the neutral statement of a single conversational turn's intent,
// composed once by the application and rendered per platform (architecture.md
// §4.1).
//
// Text is required; everything else is optional. The optional parts combine
// under these rules, which the app must respect and Validate enforces:
//
//   - Prompt and URLAction MAY coexist. They render together where a platform
//     supports both in one message (Telegram: inline buttons + a URL button in
//     the same keyboard). Where they cannot (WhatsApp in-window: reply buttons
//     and a cta_url button are mutually exclusive per message — capability-map
//     whatsapp/cta-url-button), the renderer emits them as two messages, prompt
//     first. The neutral layer permits the combination; the renderer resolves
//     it.
//
//   - LivePanel is an update-in-place marker; it composes with Text/Prompt (you
//     can edit a panel that has buttons). It has no meaning together with
//     Proactive on WhatsApp (an out-of-window proactive send is a fresh template,
//     never an edit) — Validate rejects that combination as incoherent.
//
//   - Proactive nil means "reply"; non-nil means "bot-initiated". A reply can
//     never be delivered out of a closed window by construction (a reply implies
//     the user just wrote, so the window is open) — the renderer relies on this.
//
//   - Media composes with any of the above; a platform without media support
//     degrades it (the renderer records the loss).
type MessagePlan struct {
	Text      Rich           // required
	Prompt    *ActionPrompt  // optional: choose one of N, ≤64-byte tokens
	URLAction *URLAction     // optional: one labelled web link
	LivePanel *LivePanel     // optional: update-this-message-if-possible
	Proactive *ProactiveSpec // nil = reply; non-nil = proactive (purpose→template)
	Media     *MediaRef      // optional: one image
}

// IsProactive reports whether the plan is a bot-initiated send (as opposed to a
// reply). Equivalent to plan.Proactive != nil.
func (p MessagePlan) IsProactive() bool { return p.Proactive != nil }

// Validate checks the plan's neutral-layer invariants and delegates to the parts.
//
// It enforces the required-Text rule, the LivePanel-with-Proactive
// incoherence, and each part's own Validate (prompt token lengths, URL action
// completeness). It does NOT check platform-specific limits (button counts,
// markup) — those belong to the renderer, which knows the target's Descriptor.
func (p MessagePlan) Validate() error {
	if p.Text.IsEmpty() && p.Media == nil {
		return fmt.Errorf("%w: plan has neither text nor media", ErrInvalidPlan)
	}
	if p.LivePanel != nil && p.Proactive != nil {
		return fmt.Errorf("%w: LivePanel (edit-in-place) cannot combine with Proactive (fresh send)", ErrInvalidPlan)
	}
	if p.Prompt != nil {
		if err := p.Prompt.Validate(); err != nil {
			return err
		}
	}
	if p.URLAction != nil {
		if err := p.URLAction.Validate(); err != nil {
			return err
		}
	}
	return nil
}
