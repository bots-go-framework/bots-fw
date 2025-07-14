package botsfw

import "github.com/bots-go-framework/bots-fw/botinput"

// Router dispatches requests to commands by input type, command code or a matching function
type Router interface {
	RegisterCommands(commands ...Command)
	RegisterCommandsForInputType(inputType botinput.Type, commands ...Command)

	// Dispatch requests to commands by input type, command code or a matching function
	Dispatch(webhookHandler WebhookHandler, responder WebhookResponder, whc WebhookContext) error

	// RegisteredCommands returns all registered commands
	RegisteredCommands() map[botinput.Type]map[CommandCode]Command
}
