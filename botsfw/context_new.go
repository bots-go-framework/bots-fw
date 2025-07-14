package botsfw

import "github.com/bots-go-framework/bots-fw/botinput"

// WebhookNewContext TODO: needs to be checked & described
type WebhookNewContext struct {
	BotContext
	botinput.InputMessage
}
