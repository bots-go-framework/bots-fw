package viber_bot

import "github.com/strongo/bots-framework/core"

type ViberWebhookEntry struct {
	ViberWebhookInput
}

var _ bots.WebhookEntry = (*ViberWebhookEntry)(nil)

func (whe ViberWebhookEntry) GetID() interface{} {
	panic("Not implemented")
}
