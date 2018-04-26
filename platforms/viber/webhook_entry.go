package viber

import "github.com/strongo/bots-framework/core"

// WebhookEntry base struct for Viber entities
type WebhookEntry struct {
	webhookInput
}

var _ bots.WebhookEntry = (*WebhookEntry)(nil)

// GetID is not implemented
func (whe WebhookEntry) GetID() interface{} {
	panic("Not implemented")
}
