package botsfw

import (
	"errors"
	"fmt"
)

// ErrCallbackQueryAcknowledgementUnsupported is returned when the current
// platform adapter cannot acknowledge callback queries before a handler
// returns.
var ErrCallbackQueryAcknowledgementUnsupported = errors.New("callback query acknowledgement is not supported")

// CallbackQueryAcknowledger is implemented by platform webhook contexts that
// can dismiss a callback-query loading indicator immediately.
type CallbackQueryAcknowledger interface {
	AcknowledgeCallbackQuery(text string, showAlert bool) error
	WasCallbackQueryAcknowledged() bool
}

// AcknowledgeCallbackQuery immediately answers the callback query represented
// by whc. Long-running handlers should call it before doing network or AI work.
// The router observes the acknowledgement marker and does not send its normal
// fallback answer a second time.
func AcknowledgeCallbackQuery(whc WebhookContext, text string, showAlert bool) error {
	if whc == nil {
		return fmt.Errorf("%w: nil webhook context", ErrCallbackQueryAcknowledgementUnsupported)
	}
	acknowledger, ok := whc.(CallbackQueryAcknowledger)
	if !ok {
		return fmt.Errorf("%w for platform %q", ErrCallbackQueryAcknowledgementUnsupported, whc.BotPlatform().ID())
	}
	return acknowledger.AcknowledgeCallbackQuery(text, showAlert)
}

// WasCallbackQueryAcknowledged reports whether the current callback query was
// already answered explicitly by its handler.
func WasCallbackQueryAcknowledged(whc WebhookContext) bool {
	if whc == nil {
		return false
	}
	acknowledger, ok := whc.(CallbackQueryAcknowledger)
	return ok && acknowledger.WasCallbackQueryAcknowledged()
}
