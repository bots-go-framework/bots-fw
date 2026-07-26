package botsfw

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/i18n"
)

type BotCommandOrder string

const (
	// BotCommandOrderDeclared preserves the order in BotTranslations.Commands.
	BotCommandOrderDeclared BotCommandOrder = ""
	// BotCommandOrderAlphabetical orders all published commands by command code.
	BotCommandOrderAlphabetical BotCommandOrder = "alphabetical"
	// BotCommandOrderPinnedThenAlphabetical puts explicitly pinned commands
	// first and orders every remaining command by command code.
	BotCommandOrderPinnedThenAlphabetical BotCommandOrder = "pinned_then_alphabetical"
)

type BotTranslations struct {
	Description      string
	ShortDescription string
	Commands         []BotCommand
}

// OrderBotCommands returns a copy arranged according to order.
// pinnedCommands is only valid with BotCommandOrderPinnedThenAlphabetical,
// and the supplied pin order is preserved.
func OrderBotCommands(
	commands []BotCommand,
	order BotCommandOrder,
	pinnedCommands ...string,
) ([]BotCommand, error) {
	orderedCommands := slices.Clone(commands)
	switch order {
	case BotCommandOrderDeclared:
		if len(pinnedCommands) > 0 {
			return nil, errors.New("pinned commands require pinned_then_alphabetical command order")
		}
		return orderedCommands, nil

	case BotCommandOrderAlphabetical:
		if len(pinnedCommands) > 0 {
			return nil, errors.New("pinned commands require pinned_then_alphabetical command order")
		}
		sort.SliceStable(orderedCommands, func(i, j int) bool {
			return orderedCommands[i].Command < orderedCommands[j].Command
		})
		return orderedCommands, nil

	case BotCommandOrderPinnedThenAlphabetical:
		pinnedOrder, err := pinnedCommandOrder(commands, pinnedCommands)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(orderedCommands, func(i, j int) bool {
			iPinnedOrder, iIsPinned := pinnedOrder[orderedCommands[i].Command]
			jPinnedOrder, jIsPinned := pinnedOrder[orderedCommands[j].Command]
			switch {
			case iIsPinned && jIsPinned:
				return iPinnedOrder < jPinnedOrder
			case iIsPinned:
				return true
			case jIsPinned:
				return false
			default:
				return orderedCommands[i].Command < orderedCommands[j].Command
			}
		})
		return orderedCommands, nil

	default:
		return nil, fmt.Errorf("unknown bot command order: %q", order)
	}
}

func pinnedCommandOrder(commands []BotCommand, pinnedCommands []string) (map[string]int, error) {
	commandCounts := make(map[string]int, len(commands))
	for _, command := range commands {
		commandCounts[command.Command]++
	}

	pinnedOrder := make(map[string]int, len(pinnedCommands))
	for i, commandCode := range pinnedCommands {
		if _, ok := pinnedOrder[commandCode]; ok {
			return nil, fmt.Errorf("published command %q is pinned more than once", commandCode)
		}
		switch commandCounts[commandCode] {
		case 0:
			return nil, fmt.Errorf("pinned published command %q is missing", commandCode)
		case 1:
			pinnedOrder[commandCode] = i
		default:
			return nil, fmt.Errorf("pinned published command %q is duplicated", commandCode)
		}
	}
	return pinnedOrder, nil
}

type BotCommand struct {
	Command     string `json:"command"`     // Text of the command; 1-32 characters. Can contain only lowercase English letters, digits and underscores.
	Description string `json:"description"` // Description of the command; 1-256 characters.
}

func (v BotCommand) Validate() error {
	if len(v.Command) == 0 {
		return errors.New("command is required")
	}
	if len(v.Command) > 32 {
		return errors.New("command is too long, expected to be 32 characters max")
	}
	if len(v.Description) == 0 {
		return errors.New("description is required")
	}
	if len(v.Description) > 256 {
		return errors.New("description is too long, expected to be 256 characters max")
	}
	return nil
}

type BotProfile interface {
	ID() string
	Router() Router
	DefaultLocale() i18n.Locale
	SupportedLocales() []i18n.Locale
	NewBotChatData() botsfwmodels.BotChatData
	NewPlatformUserData() botsfwmodels.PlatformUserData
	GetTranslations() BotTranslations
}

type BotProfileOption func(*botProfileConfig)

type botProfileConfig struct {
	commandOrder   BotCommandOrder
	pinnedCommands []string
}

// WithPublishedCommandOrder configures how a profile exposes its published
// commands. Pinned command codes are required to occur exactly once.
func WithPublishedCommandOrder(
	order BotCommandOrder,
	pinnedCommands ...string,
) BotProfileOption {
	return func(config *botProfileConfig) {
		config.commandOrder = order
		config.pinnedCommands = slices.Clone(pinnedCommands)
	}
}

var _ BotProfile = (*botProfile)(nil)

type botProfile struct {
	id               string
	defaultLocale    i18n.Locale
	supportedLocales []i18n.Locale
	newBotChatData   func() botsfwmodels.BotChatData
	newBotUserData   func() botsfwmodels.PlatformUserData
	router           Router
	translations     BotTranslations
}

func (v *botProfile) ID() string {
	return v.id
}

func (v *botProfile) Router() Router {
	return v.router
}

func (v *botProfile) DefaultLocale() i18n.Locale {
	return v.defaultLocale
}

func (v *botProfile) SupportedLocales() []i18n.Locale {
	return v.supportedLocales[:]
}

func (v *botProfile) NewBotChatData() botsfwmodels.BotChatData {
	return v.newBotChatData()
}

func (v *botProfile) NewPlatformUserData() botsfwmodels.PlatformUserData {
	return v.newBotUserData()
}

func (v *botProfile) GetTranslations() BotTranslations {
	return v.translations
}

func NewBotProfile(
	id string,
	router Router,
	newBotChatData func() botsfwmodels.BotChatData,
	newBotUserData func() botsfwmodels.PlatformUserData,
	defaultLocale i18n.Locale,
	supportedLocales []i18n.Locale,
	translations BotTranslations,
	options ...BotProfileOption,
) BotProfile {
	if strings.TrimSpace(id) == "" {
		panic("missing required parameter: id")
	}
	if newBotChatData == nil {
		panic("missing required parameter: newBotChatData")
	}
	if newBotUserData == nil {
		panic("missing required parameter: newBotUserData")
	}
	var defaultLocaleInSupportedLocales bool
	for _, locale := range supportedLocales {
		if locale.Code5 == defaultLocale.Code5 {
			defaultLocaleInSupportedLocales = true
			break
		}
	}
	if !defaultLocaleInSupportedLocales {
		supportedLocales = append(supportedLocales, defaultLocale)
	}
	var config botProfileConfig
	for i, option := range options {
		if option == nil {
			panic(fmt.Sprintf("nil bot profile option at index %d", i))
		}
		option(&config)
	}
	orderedCommands, err := OrderBotCommands(
		translations.Commands,
		config.commandOrder,
		config.pinnedCommands...,
	)
	if err != nil {
		panic(fmt.Sprintf("invalid published command order: %v", err))
	}
	translations.Commands = orderedCommands
	return &botProfile{
		id:               id,
		router:           router,
		defaultLocale:    defaultLocale,
		supportedLocales: supportedLocales,
		newBotChatData:   newBotChatData,
		newBotUserData:   newBotUserData,
		translations:     translations,
	}
}
