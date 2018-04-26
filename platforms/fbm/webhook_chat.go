package fbm

import "github.com/strongo/bots-framework/core"

// WebhookChat provides information about current FBM chat
type WebhookChat struct {
	ID string
}

var _ bots.WebhookChat = (*WebhookChat)(nil)

// GetID returns ID of current FBM chat
func (wh WebhookChat) GetID() string {
	return wh.ID
}

// GetType returns type of bot chat, always 'private' for FBM
func (wh WebhookChat) GetType() string {
	return "private"
}

// IsGroupChat indicates if current chat is a group one. Always false for FBM
func (wh WebhookChat) IsGroupChat() bool {
	return false
}

// NewFbmWebhookChat creates a new FBM chat instance
func NewFbmWebhookChat(id string) WebhookChat {
	return WebhookChat{
		ID: id,
	}
}
