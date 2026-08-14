package botswebhook

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/analytics"
	"github.com/strongo/logus"
	"github.com/strongo/validation"
	"net/url"
	"strings"
	"time"
)

// TypeCommands container for commands
type TypeCommands struct {
	all                  []botsfw.Command
	byCode               map[botsfw.CommandCode]botsfw.Command
	explicitTextTriggers map[string]botsfw.CommandCode
}

func newTypeCommands(commandsCount int) *TypeCommands {
	return &TypeCommands{
		byCode:               make(map[botsfw.CommandCode]botsfw.Command, commandsCount),
		all:                  make([]botsfw.Command, 0, commandsCount),
		explicitTextTriggers: make(map[string]botsfw.CommandCode, commandsCount),
	}
}

func (v *TypeCommands) addCommand(command botsfw.Command, commandType botinput.Type) {
	if command.Code == "" {
		panic(fmt.Sprintf("Command %v is missing required property ByCode", command))
	}
	if _, ok := v.byCode[command.Code]; ok {
		panic(fmt.Sprintf("Duplicate command code for %v : %v", commandType, command.Code))
	}
	if commandType == botinput.TypeText {
		v.assertUniqueExplicitTextTriggers(command)
	}
	v.all = append(v.all, command)
	v.byCode[command.Code] = command
}

// assertUniqueExplicitTextTriggers prevents registration-order-dependent text
// routing. matchMessageCommands case-folds command input before testing a
// command code or its explicit aliases, so trigger ownership is case-insensitive
// too. A command may repeat its own implicit /code trigger in Commands.
func (v *TypeCommands) assertUniqueExplicitTextTriggers(command botsfw.Command) {
	if v.explicitTextTriggers == nil {
		v.explicitTextTriggers = make(map[string]botsfw.CommandCode)
	}
	triggers := append([]string{"/" + string(command.Code)}, command.Commands...)
	for _, trigger := range triggers {
		normalizedTrigger := strings.ToLower(trigger)
		if owner, exists := v.explicitTextTriggers[normalizedTrigger]; exists && owner != command.Code {
			panic(fmt.Sprintf("Duplicate explicit text command trigger %q: command %q conflicts with command %q", trigger, owner, command.Code))
		}
	}
	for _, trigger := range triggers {
		normalizedTrigger := strings.ToLower(trigger)
		v.explicitTextTriggers[normalizedTrigger] = command.Code
	}
}

var _ botsfw.Router = (*webhooksRouter)(nil)

type ErrorFooterArgs struct {
	BotProfileID string
	BotCode      string
}
type ErrorFooterTextFunc func(ctx context.Context, botContext ErrorFooterArgs) string

// webhooksRouter maps routes to commands
type webhooksRouter struct {
	commandsByType   map[botinput.Type]*TypeCommands
	fallbackHandlers map[botinput.Type]botsfw.CommandAction
	errorFooterText  func(ctx context.Context, botContext ErrorFooterArgs) string
}

func (whRouter *webhooksRouter) RegisteredCommands() map[botinput.Type]map[botsfw.CommandCode]botsfw.Command {
	var commandsByType = make(map[botinput.Type]map[botsfw.CommandCode]botsfw.Command)
	for inputType, typeCommands := range whRouter.commandsByType {
		commandsByType[inputType] = typeCommands.byCode
	}
	return commandsByType
}

// NewWebhookRouter creates new router
//
//goland:noinspection GoUnusedExportedFunction
func NewWebhookRouter(errorFooterText func(ctx context.Context, botContext ErrorFooterArgs) string) botsfw.Router {
	return &webhooksRouter{
		commandsByType:   make(map[botinput.Type]*TypeCommands),
		fallbackHandlers: make(map[botinput.Type]botsfw.CommandAction),
		errorFooterText:  errorFooterText,
	}
}

// SetFallbackHandler registers a catch-all CommandAction for the given input type.
// It fires only when no registered command matches; unlike a catch-all Matcher command
// it does not need to be registered last and cannot accidentally block other commands.
func (whRouter *webhooksRouter) SetFallbackHandler(inputType botinput.Type, action botsfw.CommandAction) {
	if whRouter.fallbackHandlers == nil {
		whRouter.fallbackHandlers = make(map[botinput.Type]botsfw.CommandAction)
	}
	whRouter.fallbackHandlers[inputType] = action
}

func (whRouter *webhooksRouter) CommandsCount() int {
	var count int
	for _, v := range whRouter.commandsByType {
		count += len(v.all)
	}
	return count
}

// AddCommandsGroupedByType adds commands grouped by input type
// Deprecated: Use RegisterCommands() instead
func (whRouter *webhooksRouter) AddCommandsGroupedByType(commandsByType map[botinput.Type][]botsfw.Command) {
	for inputType, commands := range commandsByType {
		whRouter.RegisterCommandsForInputType(inputType, commands...)
	}
}

// AddCommands adds commands to router. It  should be called just once with the current implementation of RegisterCommandsForInputType()
// Deprecated: Use RegisterCommands() instead
func (whRouter *webhooksRouter) AddCommands(commands ...botsfw.Command) {
	whRouter.RegisterCommands(commands...)
}

// RegisterCommandsForInputType adds commands for the given input type
func (whRouter *webhooksRouter) RegisterCommandsForInputType(inputType botinput.Type, commands ...botsfw.Command) {
	typeCommands, ok := whRouter.commandsByType[inputType]
	if !ok {
		typeCommands = newTypeCommands(len(commands))
		whRouter.commandsByType[inputType] = typeCommands
	} else if inputType == botinput.TypeInlineQuery {
		panic("Duplicate add of TypeInlineQuery")
	}
	if inputType == botinput.TypeInlineQuery && len(commands) > 1 {
		panic("inputType == TypeInlineQuery && len(commands) > 1")
	}
	for _, command := range commands {
		typeCommands.addCommand(command, inputType)
	}
	if inputType == botinput.TypeInlineQuery && len(typeCommands.all) > 1 {
		panic(fmt.Sprintf("inputType == TypeInlineQuery && len(typeCommands) > 1: %v", typeCommands.all[0]))
	}
}

type CommandsRegisterer interface {
	RegisterCommands(commands ...botsfw.Command)
}

var _ CommandsRegisterer = (*webhooksRouter)(nil)

type RegisterCommandsFunc func(commands ...botsfw.Command)
type RegisterCommandsForInputTypeFunc func(inputType botinput.Type, commands ...botsfw.Command)

// RegisterCommands is registering commands with router
// TODO: Either leave this one or AddCommands()
func (whRouter *webhooksRouter) RegisterCommands(commands ...botsfw.Command) {
	addCommand := func(t botinput.Type, command botsfw.Command) {
		typeCommands, ok := whRouter.commandsByType[t]
		if !ok {
			typeCommands = newTypeCommands(0)
			whRouter.commandsByType[t] = typeCommands
		}
		typeCommands.addCommand(command, t)
	}
	for _, command := range commands {
		if len(command.InputTypes) == 0 {
			if command.TextAction != nil {
				addCommand(botinput.TypeText, command)
			}
			if command.StartAction != nil && command.TextAction == nil {
				addCommand(botinput.TypeText, command)
			}
			if command.CallbackAction != nil {
				addCommand(botinput.TypeCallbackQuery, command)
			}
			if command.LocationAction != nil {
				addCommand(botinput.TypeLocation, command)
			}
			if command.ChosenInlineResultAction != nil {
				addCommand(botinput.TypeChosenInlineResult, command)
			}
			if command.PreCheckoutQueryAction != nil {
				addCommand(botinput.TypePreCheckoutQuery, command)
			}
			if command.SuccessfulPaymentAction != nil {
				addCommand(botinput.TypeSuccessfulPayment, command)
			}
			if command.RefundedPaymentAction != nil {
				addCommand(botinput.TypeRefundedPayment, command)
			}
			if command.Action != nil {
				panic(fmt.Errorf("command{Code=%v} has Action but no InputTypes", command.Code))
			}
		} else {
			var textAdded, callbackAdded, locationAdded, inlineQueryAdded, chosenInlineResultAdded bool
			for _, t := range command.InputTypes {
				addCommand(t, command)
				switch t {
				case botinput.TypeText:
					if command.TextAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no TextAction and no Action", command.Code, command.InputTypes))
					}
					textAdded = true
				case botinput.TypeCallbackQuery:
					if command.CallbackAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no CallbackAction and no Action", command.Code, command.InputTypes))
					}
					callbackAdded = true
				case botinput.TypePreCheckoutQuery:
					if command.PreCheckoutQueryAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no PreCheckoutQueryAction and no Action", command.Code, command.InputTypes))
					}
				case botinput.TypeSuccessfulPayment:
					if command.SuccessfulPaymentAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no SuccessfulPaymentAction and no Action", command.Code, command.InputTypes))
					}
				case botinput.TypeInlineQuery:
					if command.InlineQueryAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no InlineQueryAction and no Action", command.Code, command.InputTypes))
					}
					inlineQueryAdded = true
				case botinput.TypeChosenInlineResult:
					if command.ChosenInlineResultAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no ChosenInlineResultAction and no Action", command.Code, command.InputTypes))
					}
					chosenInlineResultAdded = true
				case botinput.TypeLocation:
					if command.LocationAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no LocationAction and no Action", command.Code, command.InputTypes))
					}
					locationAdded = true
				default:
					// OK
				}
			}
			if command.TextAction != nil && !textAdded {
				addCommand(botinput.TypeText, command)
			}
			if command.CallbackAction != nil && !callbackAdded {
				addCommand(botinput.TypeCallbackQuery, command)
			}
			if command.InlineQueryAction != nil && !inlineQueryAdded {
				addCommand(botinput.TypeInlineQuery, command)
			}
			if command.ChosenInlineResultAction != nil && !chosenInlineResultAdded {
				addCommand(botinput.TypeChosenInlineResult, command)
			}
			if command.LocationAction != nil && !locationAdded {
				addCommand(botinput.TypeLocation, command)
			}
		}
	}
}

var ErrNoCommandsMatched = errors.New("no commands matched")

func matchByQueryOrMatcher(whc botsfw.WebhookContext, input interface{ GetQuery() string }, commands map[botsfw.CommandCode]botsfw.Command, hasAction func(botsfw.Command) bool) (matchedCommand *botsfw.Command, queryURL *url.URL) {
	query := input.GetQuery()
	if query != "" {
		var err error // We ignore error if the query is not a valid URL
		if queryURL, err = url.Parse(query); err == nil {
			command := commands[botsfw.CommandCode(queryURL.Path)]
			if hasAction(command) {
				matchedCommand = &command
				return
			}
		}
	}
	for _, command := range commands {
		if command.Matcher != nil {
			if command.Matcher(command, whc) {
				matchedCommand = &command
				return
			}
		}
	}
	return
}

func matchCallbackCommands(whc botsfw.WebhookContext, dataText string, dataURL *url.URL, commands map[botsfw.CommandCode]botsfw.Command) (matchedCommand *botsfw.Command, err error) {
	for _, c := range commands {
		if c.Matcher != nil && c.Matcher(c, whc) {
			return &c, nil
		}
	}
	if command, ok := commands[botsfw.CommandCode(dataURL.Path)]; ok {
		return &command, nil
	}
	log.Errorf(whc.Context(), fmt.Errorf("%w: %s", ErrNoCommandsMatched, fmt.Sprintf("dataText=[%v]", dataText)).Error())
	whc.Input().LogRequest() // TODO: LogRequest() should not be part of Input?
	return nil, err
}

func (whRouter *webhooksRouter) matchNonTextCommands(
	whc botsfw.WebhookContext, awaitingReplyTo string, commands []botsfw.Command,
) (
	matchedCommand *botsfw.Command,
) {
	if awaitingReplyTo == "" {
		if len(commands) == 1 && commands[0].Code == "" {
			matchedCommand = &commands[0]
		}
		return
	}
	if i := strings.Index(awaitingReplyTo, "?"); i != -1 {
		awaitingReplyTo = awaitingReplyTo[:i]
	}
	for _, c := range commands {
		if string(c.Code) == awaitingReplyTo || c.Matcher != nil && c.Matcher(c, whc) {
			matchedCommand = &c
			return
		}
	}
	return
}

// matchMessageCommands picks the command to handle a text message, applying a
// FIXED precedence (highest first). Each tier is tried in full before the next;
// the first tier to match wins:
//
//  1. Explicit command      — "/code", "/start <code>", or a Command.Commands alias.
//  2. AwaitingReplyTo        — the chat is mid-flow (e.g. a wizard step); the reply
//     belongs to that command. This OUTRANKS the content
//     matchers below so a catch-all Matcher can never
//     hijack an in-progress flow.
//  3. Content matchers       — Command.ExactMatch, then Command.DefaultTitle, then
//     Command.Matcher (first matching command, in
//     registration order).
//  4. No match               — returns nil; the caller then tries the input-type
//     fallback handler (SetFallbackHandler) and finally
//     WebhookHandler.HandleUnmatched.
//
// Precedence is by TIER, not registration order: a later-registered awaiting-reply
// command still beats an earlier catch-all Matcher. Set RoutingTracer to observe
// which tier/command won for a given message.
func (whRouter *webhooksRouter) matchMessageCommands(
	whc botsfw.WebhookContext, input botinput.Message, isCommandText bool, messageText, parentPath string, commands []botsfw.Command,
) (
	matchedCommand *botsfw.Command,
) {
	c := whc.Context()

	messageTextLowerCase := strings.ToLower(messageText)

	// if parentPath == "" {
	// 	log.Debugf(c, "matchMessageCommands()")
	// }

	var awaitingReplyTo string

	if !isCommandText {
		chatEntity := whc.ChatData()
		awaitingReplyTo = chatEntity.GetAwaitingReplyTo()
	}

	// log.Debugf(c, "awaitingReplyTo: %v", awaitingReplyTo)

	// Tier 1 — explicit commands ("/code", "/start <code>", Command.Commands alias).
	// These outrank everything, including an in-progress AwaitingReplyTo flow, so a
	// user can always e.g. "/cancel" out of a wizard.
	{
		commandText := messageTextLowerCase
		if atIndex := strings.Index(commandText, "@"); isCommandText && atIndex >= 0 {
			commandText = commandText[:atIndex]
		}

		var startText string
		const startWithParamsPrefixLen = len("/start ")
		if len(commandText) > startWithParamsPrefixLen && strings.HasPrefix(commandText, "/start ") {
			startText = commandText[startWithParamsPrefixLen:]
		}

		var startCommand *botsfw.Command

		for _, command := range commands {
			if isCommandText {
				if commandText == "/"+string(command.Code) || strings.HasPrefix(commandText, "/"+string(command.Code)+" ") {
					log.Debugf(c, "command matched by command.Code=%s", command.Code)
					if startText != "" {
						startCommand = &command
						continue
					} else {
						traceRoute(c, string(command.Code), "command-code", awaitingReplyTo, messageText)
						matchedCommand = &command
						return
					}
				}
				if startText != "" && command.StartAction != nil {
					if startText == string(command.Code) {
						traceRoute(c, string(command.Code), "start-command", awaitingReplyTo, messageText)
						matchedCommand = &command
						return
					}
				}
			}
			for _, commandName := range command.Commands {
				if commandName == commandText || strings.HasPrefix(messageTextLowerCase, commandName+" ") {
					log.Debugf(c, "command(code=%v) matched by command.commands", command.Code)
					traceRoute(c, string(command.Code), "command-alias", awaitingReplyTo, messageText)
					matchedCommand = &command
					return
				}
			}
		}
		if startCommand != nil {
			traceRoute(c, string(startCommand.Code), "start-command", awaitingReplyTo, messageText)
			matchedCommand = startCommand
			return
		}
	}

	// Tier 2 — AwaitingReplyTo. A reply the user is giving to an in-progress flow
	// (e.g. a wizard step armed via chatData.SetAwaitingReplyTo) MUST take priority
	// over the Tier 3 content matchers below (ExactMatch / DefaultTitle / Matcher),
	// so a catch-all Matcher command can never hijack it. Only the Tier 1 explicit
	// "/commands" above outrank an awaiting reply. Running this as a dedicated pass
	// (before Tier 3) makes the priority explicit and independent of registration
	// order — a later-registered awaiting command still beats an earlier Matcher.
	if awaitingReplyTo != "" {
		for _, command := range commands {
			awaitingReplyPrefix := strings.TrimLeft(parentPath+botsfwmodels.AwaitingReplyToPathSeparator+string(command.Code), botsfwmodels.AwaitingReplyToPathSeparator)
			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				if matched := whRouter.matchMessageCommands(whc, input, isCommandText, messageText, awaitingReplyPrefix, command.Replies); matched != nil {
					log.Debugf(c, "%v matched by command.replies", command.Code)
					traceRoute(c, string(command.Code), "awaiting-reply-replies", awaitingReplyTo, messageText)
					return matched
				}
			}
			awaitingReplyToPath := botsfwmodels.AwaitingReplyToPath(awaitingReplyTo)
			if awaitingReplyToPath == string(command.Code) || strings.HasSuffix(awaitingReplyToPath, botsfwmodels.AwaitingReplyToPathSeparator+string(command.Code)) {
				log.Debugf(c, "%v matched by awaitingReplyTo path", command.Code)
				traceRoute(c, string(command.Code), "awaiting-reply", awaitingReplyTo, messageText)
				matched := command
				return &matched
			}
		}
	}

	// Tier 3 — content matchers (ExactMatch, then DefaultTitle, then Matcher).
	// Reached only when no Tier 1/2 command claimed the message, so a wizard step is
	// never overridden by a matcher. First matching command wins, in registration order.
	for _, command := range commands {
		if command.ExactMatch != "" && (command.ExactMatch == messageText || whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			log.Debugf(c, "%v matched by command.exactMatch", command.Code)
			traceRoute(c, string(command.Code), "exact-match", awaitingReplyTo, messageText)
			matched := command
			return &matched
		}
		if command.DefaultTitle(whc) == messageText {
			log.Debugf(c, "%v matched by command.DefaultTitle", command.Code)
			traceRoute(c, string(command.Code), "default-title", awaitingReplyTo, messageText)
			matched := command
			return &matched
		}
		if command.Matcher != nil && command.Matcher(command, whc) {
			log.Debugf(c, "%v matched by command.matcher()", command.Code)
			traceRoute(c, string(command.Code), "matcher", awaitingReplyTo, messageText)
			matched := command
			return &matched
		}
	}

	traceRoute(c, "", "no-match", awaitingReplyTo, messageText)
	return nil
}

// DispatchInlineQuery dispatches inlines query
func (whRouter *webhooksRouter) DispatchInlineQuery(responder botsfw.WebhookResponder) {
	panic(fmt.Errorf("not implemented, responder: %+v", responder))
}

func changeLocaleIfLangPassed(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botmsg.MessageFromBot, err error) {
	c := whc.Context()
	q := callbackUrl.Query()
	lang := q.Get("l")
	if len(lang) == 2 {
		lang = lang + "-" + strings.ToUpper(lang)
	}
	switch lang {
	case "":
		// No language selected, for example back from submenu
	case "en-EN":
		lang = "en-US" //
	case "fa-FA":
		lang = "fa-IR" //
	default:
		//if len(lang) != 5 {
		//	m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
		//		TypeText: "Unknown language: " + lang,
		//	})
		//	log.Errorf(whc.Context(), "Unknown language: "+lang)
		//	return
		//}
	}
	if lang != "" {
		chatEntity := whc.ChatData() // We need it to be loaded before changing current Locale
		currentLang := q.Get("cl")
		currentLocaleCode5 := whc.Locale().Code5
		log.Debugf(whc.Context(), "query: %v, lang: %v, currentLang: %v, currentLocaleCode5: %v", q, lang, currentLang, currentLocaleCode5)
		if lang != currentLocaleCode5 {
			if err = whc.SetLocale(lang); err != nil {
				log.Errorf(c, "Failed to set current Locale to %v: %v", lang, err)
				err = nil
			} else {
				if currentLocaleCode5 = whc.Locale().Code5; currentLocaleCode5 != lang {
					log.Errorf(c, "DefaultLocale not set, expected %v, got: %v", lang, currentLocaleCode5)
				}
				chatEntity.SetPreferredLanguage(lang)
			}
		}
		//if lang == currentLang {
		//	m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
		//		TypeText: "It is already current language",
		//	})
		//	return
		//}
	}
	return
}

// Dispatch a query to commands
func (whRouter *webhooksRouter) Dispatch(webhookHandler botsfw.WebhookHandler, responder botsfw.WebhookResponder, whc botsfw.WebhookContext) (err error) {
	ctx := whc.Context()
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Criticalf(ctx, "*webhooksRouter.Dispatch() => PANIC: %v", err)
	// 	}
	// }()

	input := whc.Input()

	inputType := input.InputType()

	typeCommands, found := whRouter.commandsByType[inputType]
	if !found {
		// Before giving up, try an explicit fallback registered for this input type.
		if fallback, ok := whRouter.fallbackHandlers[inputType]; ok {
			var m botmsg.MessageFromBot
			m, err = fallback(whc)
			if whRouter.processCommandResponse(nil, responder, whc, m, err) {
				err = nil
			}
			return
		}
		log.Debugf(ctx, "No commands found to match by inputType: %v", botinput.GetBotInputTypeIdNameString(inputType))
		whc.Input().LogRequest()
		logInputDetails(whc, false)
		return
	}

	var (
		matchedCommand *botsfw.Command
		commandAction  botsfw.CommandAction
		m              botmsg.MessageFromBot
	)

	if len(typeCommands.all) == 0 {
		panic("len(typeCommands.all) == 0")
	}

	var isInlineQuery bool

	switch input := input.(type) {
	case botinput.CallbackQuery:
		callbackData := input.GetData()
		var callbackURL *url.URL
		if callbackData != "" {
			if callbackURL, err = url.Parse(callbackData); err != nil {
				log.Warningf(whc.Context(), "Failed to parse callback data to URL: %v", err.Error())
			}
		}
		matchedCommand, err = matchCallbackCommands(whc, callbackData, callbackURL, typeCommands.byCode)
		if err == nil && matchedCommand != nil {
			if matchedCommand.Code == "" {
				err = fmt.Errorf("matchedCommand(%T: %v).ByCode is empty string", matchedCommand, matchedCommand)
			} else if matchedCommand.CallbackAction == nil {
				err = fmt.Errorf("matchedCommand(%T: %v).CallbackAction == nil", matchedCommand, matchedCommand.Code)
			} else {
				log.Debugf(ctx, "matchCallbackCommands() => matchedCommand: %T(code=%v)", matchedCommand, matchedCommand.Code)
				if m, err = changeLocaleIfLangPassed(whc, callbackURL); err != nil || m.Text != "" {
					return
				}
				commandAction = func(whc botsfw.WebhookContext) (botmsg.MessageFromBot, error) {
					return matchedCommand.CallbackAction(whc, callbackURL)
				}
			}
		}
	case botinput.InlineQuery:
		isInlineQuery = true
		var queryURL *url.URL
		if matchedCommand, queryURL = matchByQueryOrMatcher(whc, input, typeCommands.byCode, func(command botsfw.Command) bool {
			return command.InlineQueryAction != nil || command.Action != nil
		}); matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0] // TODO: fallback to default command
		}
		if matchedCommand != nil {
			if matchedCommand.InlineQueryAction == nil {
				commandAction = matchedCommand.Action
			} else {
				commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
					return matchedCommand.InlineQueryAction(whc, input, queryURL)
				}
			}
		}
	case botinput.ChosenInlineResult:
		var queryURL *url.URL

		if matchedCommand, queryURL = matchByQueryOrMatcher(whc, input, typeCommands.byCode, func(command botsfw.Command) bool {
			return command.ChosenInlineResultAction != nil || command.Action != nil
		}); matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0] // TODO: fallback to default command
		}
		if matchedCommand == nil {
			log.Debugf(ctx, "No command found for ChosenInlineResult")
			return nil
		}
		if matchedCommand.ChosenInlineResultAction == nil {
			commandAction = matchedCommand.Action
		} else {
			commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
				return matchedCommand.ChosenInlineResultAction(whc, input, queryURL)
			}
		}
	case botinput.TextMessage:
		messageText := input.Text()
		isCommandText := strings.HasPrefix(messageText, "/")
		matchedCommand = whRouter.matchMessageCommands(whc, input, isCommandText, messageText, "", typeCommands.all)
		if matchedCommand != nil {
			if isCommandText && strings.HasPrefix(messageText, "/start") && matchedCommand.StartAction != nil {
				commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
					return matchedCommand.StartAction(whc, messageText)
				}
			} else if matchedCommand.TextAction == nil {
				commandAction = matchedCommand.Action
			} else {
				commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
					return matchedCommand.TextAction(whc, messageText)
				}
			}
		}
	case botinput.PreCheckoutQuery:
		payloadData := input.GetInvoicePayload()
		var payloadURL *url.URL
		if payloadURL, err = url.Parse(payloadData); err != nil {
			logus.Debugf(ctx, "failed to parse InvoicePayload as URL: %w", err)
			return
		}
		matchedCommand, err = matchCallbackCommands(whc, payloadData, payloadURL, typeCommands.byCode)
		if matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0]
		}
		if matchedCommand.PreCheckoutQueryAction != nil {
			commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
				return matchedCommand.PreCheckoutQueryAction(whc, input)
			}
		} else if matchedCommand.Action != nil {
			commandAction = matchedCommand.Action
		} else {
			err = fmt.Errorf("matchedCommand(code=%s) has no PreCheckoutQueryAction or Action", matchedCommand.Code)
			return
		}
	case botinput.SuccessfulPayment:
		payloadData := input.GetInvoicePayload()
		var payloadURL *url.URL
		if payloadURL, err = url.Parse(payloadData); err != nil {
			logus.Debugf(ctx, "failed to parse InvoicePayload as URL: %w", err)
			return
		}
		matchedCommand, err = matchCallbackCommands(whc, payloadData, payloadURL, typeCommands.byCode)
		if matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0]
		}
		if matchedCommand.SuccessfulPaymentAction != nil {
			commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
				return matchedCommand.SuccessfulPaymentAction(whc, input)
			}
		} else if matchedCommand.Action != nil {
			commandAction = matchedCommand.Action
		} else {
			err = fmt.Errorf("matchedCommand(code=%s) has no SuccessfulPaymentAction or Action", matchedCommand.Code)
			return
		}

	case botinput.LocationMessage:
		awaitingReplyTo := whc.ChatData().GetAwaitingReplyTo()
		matchedCommand = whRouter.matchNonTextCommands(whc, awaitingReplyTo, typeCommands.all)
		if matchedCommand != nil {
			if matchedCommand.LocationAction != nil {
				commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
					return matchedCommand.LocationAction(whc, input.GetLatitude(), input.GetLongitude())
				}
			} else if matchedCommand.Action == nil {
				commandAction = matchedCommand.Action
			} else if matchedCommand.TextAction != nil {
				commandAction = func(whc botsfw.WebhookContext) (m botmsg.MessageFromBot, err error) {
					return matchedCommand.TextAction(whc, "")
				}
			}
		}

	case botinput.Message:
		if len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0]
		} else if matchedCommand == nil {
			for _, command := range typeCommands.all {
				if command.Matcher != nil && command.Matcher(command, whc) {
					matchedCommand = &command
					break
				}
			}
		}
		if matchedCommand != nil {
			commandAction = matchedCommand.Action
		}
	default:
		if inputType == botinput.TypeUnknown {
			panic("Unknown input type")
		}
		matchedCommand = &typeCommands.all[0] // This does not feels right
		commandAction = matchedCommand.Action
	}
	if err != nil {
		err = fmt.Errorf("failed to process input{type=%s} by command{code=%s}: %w",
			botinput.GetBotInputTypeIdNameString(whc.Input().InputType()), matchedCommand.Code, err)
		if whRouter.processCommandResponseError(whc, matchedCommand, responder, err) {
			err = nil
		}
		return
	}

	if matchedCommand == nil {
		log.Debugf(ctx, "whr.matchMessageCommands() => matchedCommand == nil")
		if inputType == botinput.TypeChosenInlineResult {
			return
		}
		// Try the explicit fallback handler before falling through to HandleUnmatched.
		if fallback, ok := whRouter.fallbackHandlers[inputType]; ok {
			m, err = fallback(whc)
			if whRouter.processCommandResponse(nil, responder, whc, m, err) {
				err = nil
			}
			return
		}
		whc.Input().LogRequest()
		if m = webhookHandler.HandleUnmatched(whc); m.Text != "" || m.BotMessage != nil {
			whRouter.processCommandResponse(matchedCommand, responder, whc, m, nil)
			return
		}
		if chat := whc.Input().Chat(); chat != nil && chat.IsGroupChat() {
			// m = MessageFromBot{TypeText: "@" + whc.GetBotCode() + ": " + whc.Translate(MessageTextBotDidNotUnderstandTheCommand), Format: FormatHTML}
			// whr.processCommandResponse(matchedCommand, responder, whc, m, nil)
		} else if !isInlineQuery {
			m = whc.NewMessageByCode(botsfw.MessageTextBotDidNotUnderstandTheCommand)
			chatEntity := whc.ChatData()
			if chatEntity != nil {
				if awaitingReplyTo := chatEntity.GetAwaitingReplyTo(); awaitingReplyTo != "" {
					m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %s</i>", awaitingReplyTo)
				}
			}
			log.Debugf(ctx, "No command found for the input message: %v", whc.Input().InputType())
			whRouter.processCommandResponse(matchedCommand, responder, whc, m, nil)
		}
	} else { // matchedCommand != nil
		if matchedCommand.Code == "" {
			log.Debugf(ctx, "Matched to %T: %+v", matchedCommand, matchedCommand)
		} else {
			log.Debugf(ctx, "Matched to %T{Code=%s}", matchedCommand, matchedCommand.Code) // runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
		}
		if commandAction == nil {
			err = fmt.Errorf("no action for matched command %T{Code=%s}", matchedCommand, matchedCommand.Code)
		} else {
			m, err = commandAction(whc)
			// awaitingReplyToAfter := chatData.GetAwaitingReplyTo()
			// if isCommandText && awaitingReplyToAfter == awaitingReplyToBefore { // TODO: Looks dangerous? Should be commands be responsible?
			// 	log.Debugf(ctx, "Auto-resetting AwaitingReplyTo when not changed after processing and isCommandText=true")
			// 	chatData.SetAwaitingReplyTo("")
			// }
		}
		if err == nil {
			if chatData := whc.ChatData(); chatData != nil {
				if chatData.IsChanged() || chatData.HasChangedVars() {
					now := time.Now()
					chatData.SetDtLastInteraction(now)
					chatData.SetUpdatedTime(now)
					if err = whc.SaveBotChat(); err != nil {
						log.Errorf(ctx, "Failed to save botChat data: %v", err)
						if _, sendErr := whc.Responder().SendMessage(ctx, whc.NewMessage("Failed to save botChat data: "+err.Error()), botsfw.BotAPISendMessageOverHTTPS); sendErr != nil {
							log.Errorf(ctx, "Failed to send error message to user: %v", sendErr)
						}
					}
				}
			}

		}
		if whRouter.processCommandResponse(matchedCommand, responder, whc, m, err) {
			err = nil
		}
	}
	return
}

func logInputDetails(whc botsfw.WebhookContext, isKnownType bool) {
	c := whc.Context()
	inputType := whc.Input().InputType()
	input := whc.Input()
	inputTypeIdName := botinput.GetBotInputTypeIdNameString(inputType)
	logMessage := fmt.Sprintf("webhooksRouter.Dispatch() => WebhookIputType=%s, %T", inputTypeIdName, input)
	switch inputType {
	case botinput.TypeText:
		textMessage := input.(botinput.TextMessage)
		logMessage += fmt.Sprintf("message text: [%s]", textMessage.Text())
		if textMessage.IsEdited() { // TODO: Should be in app logic, move out of botsfw
			m := whc.NewMessage("🙇 Sorry, editing messages is not supported. Please send a new message.")
			log.Warningf(c, "TODO: Edited messages are not supported by framework yet. Move check to app.")
			_, err := whc.Responder().SendMessage(c, m, botsfw.BotAPISendMessageOverResponse)
			if err != nil {
				log.Errorf(c, "failed to send message: %v", err)
			}
			return
		}
	case botinput.TypeContact:
		contact := input.(botinput.ContactMessage)
		contactFirstName := contact.GetFirstName()
		contactBotUserID := contact.GetBotUserID()
		logMessage += fmt.Sprintf("contact number: {UserID: %s, FirstName: %s}", contactBotUserID, contactFirstName)
	case botinput.TypeInlineQuery:
		logMessage += fmt.Sprintf("inline query: [%s]", input.(botinput.InlineQuery).GetQuery())
	case botinput.TypeCallbackQuery:
		logMessage += fmt.Sprintf("callback data: [%s]", input.(botinput.CallbackQuery).GetData())
	case botinput.TypeChosenInlineResult:
		chosenResult := input.(botinput.ChosenInlineResult)
		logMessage += fmt.Sprintf("ChosenInlineResult: ResultID=[%s], InlineMessageID=[%s], Query=[%s]", chosenResult.GetResultID(), chosenResult.GetInlineMessageID(), chosenResult.GetQuery())
	case botinput.TypeReferral:
		referralMessage := input.(botinput.ReferralMessage)
		logMessage += fmt.Sprintf("referralMessage: Type=[%s], Source=[%s], Ref=[%s]", referralMessage.Type(), referralMessage.Source(), referralMessage.RefData())
	default:
		logMessage += "Unknown Type=" + botinput.GetBotInputTypeIdNameString(inputType)
	}
	if isKnownType {
		log.Debugf(c, logMessage)
	} else {
		log.Warningf(c, logMessage)
	}

	m := whc.NewMessage(fmt.Sprintf("Unknown Type=%d", inputType)) // TODO: Move out of framework to app?
	if _, err := botsfw.SendMessageThroughGate(c, whc.Responder(), m, botsfw.BotAPISendMessageOverResponse); err != nil {
		if botsfw.IsSendNotPermitted(err) {
			// Expected on gated platforms, not a failure: the platform does not
			// permit an unsolicited message right now.
			log.Warningf(c, "Not reporting unknown input type to the user: %v", err)
		} else {
			log.Errorf(c, "Failed to send message: %v", err)
		}
	}
}

func (whRouter *webhooksRouter) processCommandResponse(
	matchedCommand *botsfw.Command,
	responder botsfw.WebhookResponder,
	whc botsfw.WebhookContext,
	m botmsg.MessageFromBot,
	err error,
) (errorReportedToUser bool) {
	if err != nil {
		return whRouter.processCommandResponseError(whc, matchedCommand, responder, err)
	}

	c := whc.Context()
	if matchedCommand != nil {
		// Router-return ownership comes from the command selected by the router;
		// feature code cannot self-assert host presentation authority in a message.
		responder = botsfw.ResponseResponderForCommand(responder, matchedCommand.Code)
	}

	responseChannel := m.ResponseChannel
	if responseChannel == "" {
		responseChannel = defaultResponseChannel(c)
	}
	if gateErr := botsfw.CanSend(c, responder, m); gateErr != nil {
		// The platform does not permit this send right now — e.g. WhatsApp outside
		// the 24h customer-service window. Skip it rather than spending an API call
		// to earn a rejection. Not an error: it is the platform working as designed.
		log.Warningf(c, "command response not sent: %v", gateErr)
	} else if _, err = responder.SendMessage(c, m, responseChannel); err != nil {
		const failedToSendMessageToMessenger = "failed to send a message to messenger"
		errText := err.Error()
		switch {
		case strings.Contains(errText, "message is not modified"): // TODO: This checks are specific to Telegram and should be abstracted or moved to TG related package
			logText := failedToSendMessageToMessenger
			if whc.Input().InputType() == botinput.TypeCallbackQuery {
				logText += "(can be duplicate callback)"
			}
			log.Warningf(c, fmt.Errorf("%s: %w", logText, err).Error()) // TODO: Think how to get rid of warning on duplicate callbacks when users clicks multiple times
			err = nil
		case strings.Contains(errText, "message to edit not found"):
			log.Warningf(c, fmt.Errorf("probably an attempt to edit old or deleted message: %w", err).Error())
			err = nil
		}
		// }
		// }
		if err != nil {
			log.Errorf(c, fmt.Errorf("%s: %w", failedToSendMessageToMessenger, err).Error()) // TODO: Decide how do we handle this
		}
	}
	// Editing a Telegram message does not acknowledge the button press. Always
	// answer a successful callback unless the command supplied its own callback
	// answer, otherwise Telegram leaves the client's "Loading…" indicator open.
	if whc.Input().InputType() == botinput.TypeCallbackQuery &&
		(m.BotMessage == nil || m.BotMessage.BotMessageType() != botmsg.TypeCallbackAnswer) &&
		!botsfw.WasCallbackQueryAcknowledged(whc) {
		if callbackQuery, ok := whc.Input().(botinput.CallbackQuery); ok {
			acknowledgeCallbackQuery(c, responder, callbackQuery)
		} else {
			log.Errorf(c, "Callback-query input %T does not implement botinput.CallbackQuery", whc.Input())
		}
	}
	if matchedCommand != nil {
		//if inputType := whc.Input().InputType(); inputType != botinput.TypeCallbackQuery {
		//	chatData := whc.ChatData()
		//	if chatData != nil {
		//		path = chatData.GetAwaitingReplyTo()
		//		if path == "" {
		//			path = string(matchedCommand.Code)
		//		} else if pathURL, err := url.Parse(path); err == nil {
		//			path = pathURL.Path
		//		}
		//		title = matchedCommand.Title
		//	} else {
		//		path = botinput.GetBotInputTypeIdNameString(inputType)
		//		title = matchedCommand.Title
		//	}
		//}

		var am analytics.Message

		botCode := whc.GetBotCode()

		var pageview analytics.Pageview

		if m.Analytics != nil {
			am = m.Analytics
			pageview, _ = am.(analytics.Pageview)
		}
		path := string(matchedCommand.Code)
		if path != "" && (m.Text != "" || m.BotMessage != nil && m.BotMessage.BotMessageType() == botmsg.TypeText) {
			if path[0] != '/' {
				path = "/" + path
			}
			botPlatformID := whc.BotPlatform().ID()
			if pageview == nil || pageview.Host() != botPlatformID || pageview.Path() == "" {
				originalPageview := pageview
				if !strings.HasPrefix(path, "/bot/") {
					path = "/bot/" + path
				}
				pageview = analytics.NewPageview(botPlatformID, path)
				if originalPageview == nil {
					if matchedCommand.Title != "" {
						pageview.SetTitle(matchedCommand.Title)
					}
				} else {
					if title := originalPageview.Title(); title != "" {
						pageview.SetTitle(title)
					}
					if uc := originalPageview.User(); uc != nil {
						pageview.SetUserContext(uc)
					}
					if props := pageview.Properties(); len(props) > 0 {
						for k, v := range originalPageview.Properties() {
							props.Set(k, v)
						}
					}
				}
				am = pageview
			}

			pageview.SetUserAgent(botPlatformID)

			pageview.SetURL(botPlatformID + "://" + botCode + "/" + path)

			if matchedCommand.Title != "" && pageview.Title() == "" {
				pageview.SetTitle(matchedCommand.Title)
			}
		}
		if am != nil {
			props := am.Properties()
			props.Set("bot", botCode)
			props.Set("input_type", whc.Input().InputType().String())
			whAnalytics := whc.Analytics()
			whAnalytics.Enqueue(am)
		}
	}
	return false
}

const defaultResponseChannelEnv = "BOTSFW_DEFAULT_RESPONSE_CHANNEL"

// defaultResponseChannel returns the transport for command results that do not
// explicitly choose one. The efficient webhook response is the default; set
// BOTSFW_DEFAULT_RESPONSE_CHANNEL=https to use the Bot API instead.
func defaultResponseChannel(ctx context.Context) botmsg.BotAPISendMessageChannel {
	configured := botsfw.GetEnv(ctx, defaultResponseChannelEnv)
	switch strings.ToLower(strings.TrimSpace(configured)) {
	case "", "response":
		return botsfw.BotAPISendMessageOverResponse
	case "https":
		return botsfw.BotAPISendMessageOverHTTPS
	default:
		log.Warningf(ctx, "Unknown %s value %q; using response", defaultResponseChannelEnv, configured)
		return botsfw.BotAPISendMessageOverResponse
	}
}

func acknowledgeCallbackQuery(ctx context.Context, responder botsfw.WebhookResponder, callbackQuery botinput.CallbackQuery) {
	ack := botmsg.MessageFromBot{
		BotMessage: botmsg.AnswerCallbackQuery{CallbackQueryID: callbackQuery.GetID()},
	}
	if _, err := responder.SendMessage(ctx, ack, botsfw.BotAPISendMessageOverHTTPS); err != nil {
		log.Errorf(ctx, "Failed to acknowledge callback query: %v", err)
	}
}

// processCommandResponseError returns true only when it successfully delivered
// an error to the user. Callers then consume the error so the webhook driver
// does not append a second HTTP error to the Telegram response body.
func (whRouter *webhooksRouter) processCommandResponseError(whc botsfw.WebhookContext, matchedCommand *botsfw.Command, responder botsfw.WebhookResponder, err error) bool {
	ctx := whc.Context()
	logTelegramProviderError(ctx, err, logus.Errorf)
	env := whc.GetBotSettings().Env

	if env == botsfw.EnvProduction {
		whc.Analytics().Enqueue(analytics.NewErrorMessage(err))
	}
	//inputType := whc.Input().InputType()
	switch inputType := whc.Input().InputType(); inputType {
	case botinput.TypeText, botinput.TypeContact:
		// TODO: Try to get botChat ID from user?
		var footer string
		if whRouter.errorFooterText != nil {
			args := ErrorFooterArgs{
				BotCode:      whc.GetBotCode(),
				BotProfileID: "", // TODO(help-wanted): implement!
			}
			footer = whRouter.errorFooterText(ctx, args)
		}
		m, expandable := expandableUserErrorMessage(whc, err, footer)
		if !expandable {
			m = whc.NewMessage(
				whc.Translate(botsfw.MessageTextOopsSomethingWentWrong) +
					"\n\n" +
					"💢" +
					fmt.Sprintf(" Server error - failed to process message: %v", err),
			)
			if footer != "" {
				m.Text += "\n\n" + footer
			}
		}
		if _, respErr := responder.SendMessage(ctx, m, botsfw.BotAPISendMessageOverResponse); respErr != nil {
			if expandable && isTelegramProviderRejection(respErr) {
				logTelegramProviderError(ctx, respErr, logus.Errorf)
				plainFallback := plainUserErrorFallback(m)
				if _, fallbackErr := responder.SendMessage(
					ctx,
					plainFallback,
					botsfw.BotAPISendMessageOverResponse,
				); fallbackErr == nil {
					return true
				} else {
					log.Errorf(ctx, "Failed to report a plain-text fallback error to user for command %T: %v", matchedCommand, fallbackErr)
				}
			}
			log.Errorf(ctx, "Failed to report to user a server error for command %T: %v", matchedCommand, respErr)
			return false
		}
		return true
	case botinput.TypeCallbackQuery:
		logus.Errorf(ctx, "Failed to process callback query by command{code=%s}: %v", matchedCommand.Code, inputType)
		var msg botmsg.MessageFromBot
		msg.BotMessage = botmsg.AnswerCallbackQuery{
			CallbackQueryID: whc.Input().(botinput.CallbackQuery).GetID(),
			Text:            callbackErrorText(err),
			ShowAlert:       true,
			CacheTime:       3,
		}
		if _, err = responder.SendMessage(ctx, msg, botsfw.BotAPISendMessageOverHTTPS); err != nil {
			logus.Errorf(ctx, "Failed to send callback error message to messenger: %v", err)
			return false
		}
		return true

	default:
		logus.Errorf(ctx, "Failed to process %v input by command{code=%s}: %v", inputType, matchedCommand.Code, inputType)
	}
	return false
}

// logTelegramProviderError records Telegram's structured diagnostic when a
// command response fails. TelegramProviderErrorDetailsFrom intentionally
// excludes the bot token, request URL, request parameters, and raw response
// body; Error() stays redacted for user-facing messages.
func logTelegramProviderError(ctx context.Context, err error, errorf func(context.Context, string, ...any)) bool {
	details, ok := tgbotapi.TelegramProviderErrorDetailsFrom(err)
	if !ok {
		return false
	}
	errorf(
		ctx,
		"Telegram API provider error: method=%q, error_code=%d, description=%q",
		details.Method,
		details.ErrorCode,
		details.Description,
	)
	return true
}

// callbackErrorText returns a short, user-safe explanation for a failed inline
// action. Telegram's answerCallbackQuery rejects texts longer than 200 bytes,
// and exposing err.Error() can leak internal state and user data. The complete
// cause remains in the server log and production error analytics.
func callbackErrorText(err error) string {
	if validation.IsBadRecordError(err) {
		return "⚠️ We couldn't complete this because your account data needs attention. Please contact support."
	}
	return "⚠️ We couldn't complete this action. Please try again."
}
