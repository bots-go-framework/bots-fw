package botswebhook

import (
	"context"
	"github.com/bots-go-framework/bots-fw/botinput"
	botsfw3 "github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
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
	cmd := botsfw.Command{
		Code:       "test",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
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
	cmd1 := botsfw.Command{
		Code:       "test1",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
		},
	}
	router.AddCommands(cmd1)

	// Verify the command was added
	commands := router.RegisteredCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command type, got %d", len(commands))
	}

	// Add another command
	cmd2 := botsfw.Command{
		Code:       "test2",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
		},
	}
	router.AddCommands(cmd2)

	// Verify both commands were added
	commands = router.RegisteredCommands()
	if len(commands[botinput.TypeText]) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands[botinput.TypeText]))
	}
}

func TestWebhooksRouter_RegisterCommandsForInputType(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Add a command for a specific input type
	cmd := botsfw.Command{
		Code:       "test",
		InputTypes: []botinput.Type{botinput.TypeInlineQuery},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
		},
	}
	router.RegisterCommandsForInputType(botinput.TypeInlineQuery, cmd)

	// Verify the command was added for the correct input type
	commands := router.RegisteredCommands()
	if len(commands[botinput.TypeInlineQuery]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.TypeInlineQuery, len(commands[botinput.TypeInlineQuery]))
	}
}

func TestWebhooksRouter_AddCommandsGroupedByType(t *testing.T) {
	// Create a router
	router := NewWebhookRouter(nil).(*webhooksRouter)

	// Create commands grouped by type
	cmd1 := botsfw.Command{
		Code:       "test1",
		InputTypes: []botinput.Type{botinput.TypeText},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
		},
	}
	cmd2 := botsfw.Command{
		Code:       "test2",
		InputTypes: []botinput.Type{botinput.TypeInlineQuery},
		Action: func(whc botsfw.WebhookContext) (botsfw3.MessageFromBot, error) {
			return botsfw3.MessageFromBot{}, nil
		},
	}
	commandsByType := map[botinput.Type][]botsfw.Command{
		botinput.TypeText:        {cmd1},
		botinput.TypeInlineQuery: {cmd2},
	}

	// Add the commands
	router.AddCommandsGroupedByType(commandsByType)

	// Verify the commands were added for the correct input types
	commands := router.RegisteredCommands()
	if len(commands[botinput.TypeText]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.TypeText, len(commands[botinput.TypeText]))
	}
	if len(commands[botinput.TypeInlineQuery]) != 1 {
		t.Errorf("Expected 1 command for input type %v, got %d", botinput.TypeInlineQuery, len(commands[botinput.TypeInlineQuery]))
	}
}
