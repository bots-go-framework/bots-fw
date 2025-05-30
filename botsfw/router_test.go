package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw/botinput"
	"testing"
)

func TestNewWebhookRouter(t *testing.T) {
	// Create a router with a nil error footer text function
	router := NewWebhookRouter(nil)
	if router == nil {
		t.Error("NewWebhookRouter(nil) returned nil")
	}

	// Create a router with a custom error footer text function
	errorFooterText := func(ctx context.Context, botContext ErrorFooterArgs) string {
		return "Test error footer"
	}
	router = NewWebhookRouter(errorFooterText)
	if router == nil {
		t.Error("NewWebhookRouter(errorFooterText) returned nil")
	}
}

func TestWebhooksRouter_CommandsCount(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Initially, there should be no commands
	if count := router.CommandsCount(); count != 0 {
		t.Errorf("Expected CommandsCount() to be 0, got %d", count)
	}

	// Add a command
	cmd := Command{
		Code:       "test",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputText},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	router.AddCommands(cmd)

	// Now there should be 1 command
	if count := router.CommandsCount(); count != 1 {
		t.Errorf("Expected CommandsCount() to be 1, got %d", count)
	}
}

func TestWebhooksRouter_AddCommands(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Add a command
	cmd1 := Command{
		Code:       "test1",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputText},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	router.AddCommands(cmd1)

	// Verify the command was added
	commands := router.RegisteredCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command type, got %d", len(commands))
	}

	// Add another command
	cmd2 := Command{
		Code:       "test2",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputText},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	router.AddCommands(cmd2)

	// Verify both commands were added
	commands = router.RegisteredCommands()
	if len(commands[botinput.WebhookInputText]) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands[botinput.WebhookInputText]))
	}
}

func TestWebhooksRouter_RegisterCommandsForInputType(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Add a command for a specific input type
	cmd := Command{
		Code:       "test",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputInlineQuery},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	router.RegisterCommandsForInputType(botinput.WebhookInputInlineQuery, cmd)

	// Verify the command was added for the correct input type
	commands := router.RegisteredCommands()
	if len(commands[botinput.WebhookInputInlineQuery]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.WebhookInputInlineQuery, len(commands[botinput.WebhookInputInlineQuery]))
	}
}

func TestWebhooksRouter_AddCommandsGroupedByType(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Create commands grouped by type
	cmd1 := Command{
		Code:       "test1",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputText},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	cmd2 := Command{
		Code:       "test2",
		InputTypes: []botinput.WebhookInputType{botinput.WebhookInputInlineQuery},
		Action: func(whc WebhookContext) (MessageFromBot, error) {
			return MessageFromBot{}, nil
		},
	}
	commandsByType := map[botinput.WebhookInputType][]Command{
		botinput.WebhookInputText:        {cmd1},
		botinput.WebhookInputInlineQuery: {cmd2},
	}

	// Add the commands
	router.AddCommandsGroupedByType(commandsByType)

	// Verify the commands were added for the correct input types
	commands := router.RegisteredCommands()
	if len(commands[botinput.WebhookInputText]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.WebhookInputText, len(commands[botinput.WebhookInputText]))
	}
	if len(commands[botinput.WebhookInputInlineQuery]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.WebhookInputInlineQuery, len(commands[botinput.WebhookInputInlineQuery]))
	}
}
