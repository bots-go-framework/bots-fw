package botplan

import "fmt"

// ErrInvalidPlan is the base error for a MessagePlan (or one of its parts) that
// violates a neutral-layer invariant. Callers can match it with errors.Is.
var ErrInvalidPlan = fmt.Errorf("botplan: invalid message plan")

// ErrTokenTooLong is returned when a Choice.Token exceeds MaxChoiceTokenBytes.
// It wraps ErrInvalidPlan, so errors.Is(err, ErrInvalidPlan) also holds — a
// too-long token is one kind of invalid plan.
var ErrTokenTooLong = fmt.Errorf("%w: choice token exceeds 64 bytes", ErrInvalidPlan)

// ErrNoTemplateForPurpose is returned by a renderer when a proactive plan must
// be delivered out of the platform's send window but no approved template exists
// for its ProactiveSpec.Purpose.
//
// This is not a bug — it is a scenario-specific policy point (scenario-catalogue
// SYS-TPL-030: "degrade the scenario, not the product"). The renderer surfaces
// it typed so the caller can apply that scenario's declared fallback (hold until
// the window reopens, drop, defer to a digest, …) rather than the renderer
// guessing one.
var ErrNoTemplateForPurpose = fmt.Errorf("botplan: no approved template for proactive purpose")

// ErrTemplateMismatch is returned by a renderer when an approved template exists
// for the purpose but its shape does not match the plan — for example the
// prompt's choices do not line up with the template's approved quick-reply
// labels, or the params do not cover the template's body placeholders. Quick-
// reply buttons on a WhatsApp template are fixed at approval time, not composed
// per send (capability-map whatsapp/template-buttons buttonsFixedAtApprovalNotPerSend),
// so a plan that assumes different buttons cannot be honoured and must fail
// loudly rather than silently drop the choices.
var ErrTemplateMismatch = fmt.Errorf("botplan: plan does not match approved template shape")

// ErrUnsupportedTarget is returned by a renderer asked to render for a platform
// it does not implement.
var ErrUnsupportedTarget = fmt.Errorf("botplan: unsupported render target platform")
