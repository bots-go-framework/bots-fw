package botsfw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderBotCommands(t *testing.T) {
	commands := []BotCommand{
		{Command: "watch", Description: "Watch"},
		{Command: "sports", Description: "Sports"},
		{Command: "find", Description: "Find"},
		{Command: "main", Description: "Main"},
		{Command: "games", Description: "Games"},
	}

	t.Run("declared", func(t *testing.T) {
		ordered, err := OrderBotCommands(commands, BotCommandOrderDeclared)

		require.NoError(t, err)
		assert.Equal(t, []string{"watch", "sports", "find", "main", "games"}, botCommandCodes(ordered))
		ordered[0].Command = "changed"
		assert.Equal(t, "watch", commands[0].Command, "OrderBotCommands must return a copy")
	})

	t.Run("alphabetical", func(t *testing.T) {
		ordered, err := OrderBotCommands(commands, BotCommandOrderAlphabetical)

		require.NoError(t, err)
		assert.Equal(t, []string{"find", "games", "main", "sports", "watch"}, botCommandCodes(ordered))
	})

	t.Run("pinned_then_alphabetical", func(t *testing.T) {
		ordered, err := OrderBotCommands(
			commands,
			BotCommandOrderPinnedThenAlphabetical,
			"main",
			"sports",
		)

		require.NoError(t, err)
		assert.Equal(t, []string{"main", "sports", "find", "games", "watch"}, botCommandCodes(ordered))
	})
}

func TestOrderBotCommandsRejectsInvalidConfiguration(t *testing.T) {
	commands := []BotCommand{{Command: "main", Description: "Main"}}

	tests := []struct {
		name         string
		order        BotCommandOrder
		pinned       []string
		extraCommand []BotCommand
		errorText    string
	}{
		{
			name:      "pins_with_declared_order",
			pinned:    []string{"main"},
			errorText: "pinned commands require pinned_then_alphabetical command order",
		},
		{
			name:      "pins_with_alphabetical_order",
			order:     BotCommandOrderAlphabetical,
			pinned:    []string{"main"},
			errorText: "pinned commands require pinned_then_alphabetical command order",
		},
		{
			name:      "unknown_order",
			order:     BotCommandOrder("popular"),
			errorText: `unknown bot command order: "popular"`,
		},
		{
			name:      "missing_pin",
			order:     BotCommandOrderPinnedThenAlphabetical,
			pinned:    []string{"sports"},
			errorText: `pinned published command "sports" is missing`,
		},
		{
			name:      "duplicate_pin",
			order:     BotCommandOrderPinnedThenAlphabetical,
			pinned:    []string{"main", "main"},
			errorText: `published command "main" is pinned more than once`,
		},
		{
			name:         "duplicated_pinned_command",
			order:        BotCommandOrderPinnedThenAlphabetical,
			pinned:       []string{"main"},
			extraCommand: []BotCommand{{Command: "main", Description: "Main again"}},
			errorText:    `pinned published command "main" is duplicated`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := OrderBotCommands(
				append(commands, test.extraCommand...),
				test.order,
				test.pinned...,
			)

			require.EqualError(t, err, test.errorText)
		})
	}
}

func botCommandCodes(commands []BotCommand) []string {
	result := make([]string, len(commands))
	for i, command := range commands {
		result[i] = command.Command
	}
	return result
}
