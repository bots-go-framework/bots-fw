package fbm_bot

import "github.com/strongo/bots-framework/core"

// FbmWebhookChat provides information about current FBM chat
type FbmWebhookChat struct {
	ID string
}

var _ bots.WebhookChat = (*FbmWebhookChat)(nil)

// GetID returns ID of current FBM chat
func (wh FbmWebhookChat) GetID() string {
	return wh.ID
}

// GetType returns type of bot chat, always 'private' for FBM
func (wh FbmWebhookChat) GetType() string {
	return "private"
}

// IsGroupChat indicates if current chat is a group one. Always false for FBM
func (wh FbmWebhookChat) IsGroupChat() bool {
	return false
}

// NewFbmWebhookChat creates a new FBM chat instance
func NewFbmWebhookChat(id string) FbmWebhookChat {
	return FbmWebhookChat{
		ID: id,
	}
}
