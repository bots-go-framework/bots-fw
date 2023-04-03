package botsfw

import "context"

// BotChatStore is interface for DAL to store bot chat data
type BotChatStore interface {
	GetBotChatEntityByID(c context.Context, botID, botChatID string) (BotChat, error)
	SaveBotChat(c context.Context, botID, botChatID string, chatEntity BotChat) error
	NewBotChatEntity(c context.Context, botID string, botChat WebhookChat, appUserID, botUserID string, isAccessGranted bool) BotChat
	Close(c context.Context) error // TODO: Was io.Closer, should it?
}
