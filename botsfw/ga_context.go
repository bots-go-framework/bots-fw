package botsfw

import "github.com/strongo/gamp"

// GaContext provides context to Google Analytics - TODO: we should have an abstraction for analytics
type GaContext interface {
	GaQueuer
	// Flush() error
	GaCommon() gamp.Common
	GaEvent(category, action string) *gamp.Event
	GaEventWithLabel(category, action, label string) *gamp.Event
}
