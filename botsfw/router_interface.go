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

	// SetFallbackHandler registers a catch-all action for the given input type.
	// The fallback fires only when no registered command matches the input.
	// Unlike a catch-all Matcher command, the fallback is order-independent:
	// it is stored separately and never blocks the normal command-matching loop.
	// Only one fallback per input type is supported; a second call replaces the first.
	SetFallbackHandler(inputType botinput.Type, action CommandAction)
}
