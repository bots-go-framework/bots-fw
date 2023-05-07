package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw-models/botsfwmodels"
)

// BotChatStore is interface for DAL to store bot chat data
type BotChatStore interface {

	// GetBotChatEntityByID returns bot chat record by IDs
	GetBotChatEntityByID(c context.Context, botID, botChatID string) (botsfwmodels.BotChat, error)

	// SaveBotChat saves bot chat record
	SaveBotChat(c context.Context, botID, botChatID string, chatEntity botsfwmodels.BotChat) error

	// NewBotChatEntity creates new bot chat record
	NewBotChatEntity(c context.Context, botID string, botChat WebhookChat, appUserID, botUserID string, isAccessGranted bool) botsfwmodels.BotChat

	// Close closes the store, e.g. commits sends a signal to commit transaction
	// TODO: Consider to remove this method if possible
	Close(c context.Context) error // TODO: Was io.Closer, should it?
}
