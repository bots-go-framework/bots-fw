package botsfw

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/strongo/analytics"
	"github.com/strongo/logus"
	"net/url"
	"strings"
	"time"
)

// TypeCommands container for commands
type TypeCommands struct {
	all    []Command
	byCode map[CommandCode]Command
}

func newTypeCommands(commandsCount int) *TypeCommands {
	return &TypeCommands{
		byCode: make(map[CommandCode]Command, commandsCount),
		all:    make([]Command, 0, commandsCount),
	}
}

func (v *TypeCommands) addCommand(command Command, commandType botinput.WebhookInputType) {
	if command.Code == "" {
		panic(fmt.Sprintf("Command %v is missing required property ByCode", command))
	}
	v.all = append(v.all, command)
	if _, ok := v.byCode[command.Code]; !ok {
		v.byCode[command.Code] = command
	} else {
		panic(fmt.Sprintf("Duplicate command code for %v : %v", commandType, command.Code))
	}
}

// Router dispatches requests to commands by input type, command code or a matching function
type Router interface {
	RegisterCommands(commands ...Command)
	RegisterCommandsForInputType(inputType botinput.WebhookInputType, commands ...Command)

	// Dispatch requests to commands by input type, command code or a matching function
	Dispatch(webhookHandler WebhookHandler, responder WebhookResponder, whc WebhookContext) error

	// RegisteredCommands returns all registered commands
	RegisteredCommands() map[botinput.WebhookInputType]map[CommandCode]Command
}

var _ Router = (*webhooksRouter)(nil)

type ErrorFooterArgs struct {
	BotProfileID string
	BotCode      string
}
type ErrorFooterTextFunc func(ctx context.Context, botContext ErrorFooterArgs) string

// webhooksRouter maps routes to commands
type webhooksRouter struct {
	commandsByType  map[botinput.WebhookInputType]*TypeCommands
	errorFooterText func(ctx context.Context, botContext ErrorFooterArgs) string
}

func (whRouter *webhooksRouter) RegisteredCommands() map[botinput.WebhookInputType]map[CommandCode]Command {
	var commandsByType = make(map[botinput.WebhookInputType]map[CommandCode]Command)
	for inputType, typeCommands := range whRouter.commandsByType {
		commandsByType[inputType] = typeCommands.byCode
	}
	return commandsByType
}

// NewWebhookRouter creates new router
//
//goland:noinspection GoUnusedExportedFunction
func NewWebhookRouter(errorFooterText func(ctx context.Context, botContext ErrorFooterArgs) string) Router {
	return &webhooksRouter{
		commandsByType:  make(map[botinput.WebhookInputType]*TypeCommands),
		errorFooterText: errorFooterText,
	}
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
func (whRouter *webhooksRouter) AddCommandsGroupedByType(commandsByType map[botinput.WebhookInputType][]Command) {
	for inputType, commands := range commandsByType {
		whRouter.RegisterCommandsForInputType(inputType, commands...)
	}
}

// AddCommands adds commands to router. It  should be called just once with the current implementation of RegisterCommandsForInputType()
// Deprecated: Use RegisterCommands() instead
func (whRouter *webhooksRouter) AddCommands(commands ...Command) {
	whRouter.RegisterCommands(commands...)
}

// RegisterCommandsForInputType adds commands for the given input type
func (whRouter *webhooksRouter) RegisterCommandsForInputType(inputType botinput.WebhookInputType, commands ...Command) {
	typeCommands, ok := whRouter.commandsByType[inputType]
	if !ok {
		typeCommands = newTypeCommands(len(commands))
		whRouter.commandsByType[inputType] = typeCommands
	} else if inputType == botinput.WebhookInputInlineQuery {
		panic("Duplicate add of WebhookInputInlineQuery")
	}
	if inputType == botinput.WebhookInputInlineQuery && len(commands) > 1 {
		panic("inputType == WebhookInputInlineQuery && len(commands) > 1")
	}
	for _, command := range commands {
		typeCommands.addCommand(command, inputType)
	}
	if inputType == botinput.WebhookInputInlineQuery && len(typeCommands.all) > 1 {
		panic(fmt.Sprintf("inputType == WebhookInputInlineQuery && len(typeCommands) > 1: %v", typeCommands.all[0]))
	}
}

type CommandsRegisterer interface {
	RegisterCommands(commands ...Command)
}

var _ CommandsRegisterer = (*webhooksRouter)(nil)

type RegisterCommandsFunc func(commands ...Command)
type RegisterCommandsForInputTypeFunc func(inputType botinput.WebhookInputType, commands ...Command)

// RegisterCommands is registering commands with router
// TODO: Either leave this one or AddCommands()
func (whRouter *webhooksRouter) RegisterCommands(commands ...Command) {
	addCommand := func(t botinput.WebhookInputType, command Command) {
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
				addCommand(botinput.WebhookInputText, command)
			}
			if command.StartAction != nil && command.TextAction == nil {
				addCommand(botinput.WebhookInputText, command)
			}
			if command.CallbackAction != nil {
				addCommand(botinput.WebhookInputCallbackQuery, command)
			}
			if command.ChosenInlineResultAction != nil {
				addCommand(botinput.WebhookInputChosenInlineResult, command)
			}
			if command.PreCheckoutQueryAction != nil {
				addCommand(botinput.WebhookInputPreCheckoutQuery, command)
			}
			if command.SuccessfulPaymentAction != nil {
				addCommand(botinput.WebhookInputSuccessfulPayment, command)
			}
			if command.RefundedPaymentAction != nil {
				addCommand(botinput.WebhookInputRefundedPayment, command)
			}
			if command.Action != nil {
				panic(fmt.Errorf("command{Code=%v} has Action but no InputTypes", command.Code))
			}
		} else {
			var textAdded, callbackAdded, inlineQueryAdded, chosenInlineResultAdded bool
			for _, t := range command.InputTypes {
				addCommand(t, command)
				switch t {
				case botinput.WebhookInputText:
					if command.TextAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no TextAction and no Action", command.Code, command.InputTypes))
					}
					textAdded = true
				case botinput.WebhookInputCallbackQuery:
					if command.CallbackAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no CallbackAction and no Action", command.Code, command.InputTypes))
					}
					callbackAdded = true
				case botinput.WebhookInputPreCheckoutQuery:
					if command.PreCheckoutQueryAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no PreCheckoutQueryAction and no Action", command.Code, command.InputTypes))
					}
				case botinput.WebhookInputSuccessfulPayment:
					if command.SuccessfulPaymentAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no SuccessfulPaymentAction and no Action", command.Code, command.InputTypes))
					}
				case botinput.WebhookInputInlineQuery:
					if command.InlineQueryAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no InlineQueryAction and no Action", command.Code, command.InputTypes))
					}
					inlineQueryAdded = true
				case botinput.WebhookInputChosenInlineResult:
					if command.ChosenInlineResultAction == nil && command.Action == nil {
						panic(fmt.Errorf("command{Code=%v,InputTypes=%+v} has no ChosenInlineResultAction and no Action", command.Code, command.InputTypes))
					}
					chosenInlineResultAdded = true
				default:
					// OK
				}
			}
			if command.TextAction != nil && !textAdded {
				addCommand(botinput.WebhookInputText, command)
			}
			if command.CallbackAction != nil && !callbackAdded {
				addCommand(botinput.WebhookInputCallbackQuery, command)
			}
			if command.InlineQueryAction != nil && !inlineQueryAdded {
				addCommand(botinput.WebhookInputInlineQuery, command)
			}
			if command.ChosenInlineResultAction != nil && !chosenInlineResultAdded {
				addCommand(botinput.WebhookInputChosenInlineResult, command)
			}
		}
	}
}

var ErrNoCommandsMatched = errors.New("no commands matched")

func matchByQueryOrMatcher(whc WebhookContext, input interface{ GetQuery() string }, commands map[CommandCode]Command, hasAction func(Command) bool) (matchedCommand *Command, queryURL *url.URL) {
	query := input.GetQuery()
	if query != "" {
		var err error // We ignore error if the query is not a valid URL
		if queryURL, err = url.Parse(query); err == nil {
			command := commands[CommandCode(queryURL.Path)]
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

func matchCallbackCommands(whc WebhookContext, dataText string, dataURL *url.URL, commands map[CommandCode]Command) (matchedCommand *Command, err error) {
	for _, c := range commands {
		if c.Matcher != nil && c.Matcher(c, whc) {
			return &c, nil
		}
	}
	if command, ok := commands[CommandCode(dataURL.Path)]; ok {
		return &command, nil
	}
	log.Errorf(whc.Context(), fmt.Errorf("%w: %s", ErrNoCommandsMatched, fmt.Sprintf("dataText=[%v]", dataText)).Error())
	whc.Input().LogRequest() // TODO: LogRequest() should not be part of Input?
	return nil, err
}

func (whRouter *webhooksRouter) matchMessageCommands(
	whc WebhookContext, input botinput.WebhookMessage, isCommandText bool, messageText, parentPath string, commands []Command,
) (
	matchedCommand *Command,
) {
	c := whc.Context()

	var awaitingReplyCommand Command
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

	var awaitingReplyCommandFound bool

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

		var startCommand *Command

		for _, command := range commands {
			if isCommandText {
				if commandText == "/"+string(command.Code) || strings.HasPrefix(commandText, "/"+string(command.Code)+" ") {
					log.Debugf(c, "command matched by command.Code=%s", command.Code)
					if startText != "" {
						startCommand = &command
						continue
					} else {
						matchedCommand = &command
						return
					}
				}
				if startText != "" && command.StartAction != nil {
					if startText == string(command.Code) {
						matchedCommand = &command
						return
					}
				}
			}
			for _, commandName := range command.Commands {
				if commandName == commandText || strings.HasPrefix(messageTextLowerCase, commandName+" ") {
					log.Debugf(c, "command(code=%v) matched by command.commands", command.Code)
					matchedCommand = &command
					return
				}
			}
		}
		if startCommand != nil {
			matchedCommand = startCommand
			return
		}
	}

	for _, command := range commands {
		if !awaitingReplyCommandFound && awaitingReplyTo != "" {
			awaitingReplyPrefix := strings.TrimLeft(parentPath+botsfwmodels.AwaitingReplyToPathSeparator+string(command.Code), botsfwmodels.AwaitingReplyToPathSeparator)

			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				// log.Debugf(c, "[%v] is a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
				// log.Debugf(c, "awaitingReplyCommand: %v", command.ByCode)
				if matchedCommand = whRouter.matchMessageCommands(whc, input, isCommandText, messageText, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
					log.Debugf(c, "%v matched by command.replies", command.Code)
					awaitingReplyCommand = *matchedCommand
					awaitingReplyCommandFound = true
					continue
				}
				//} else {
				// log.Debugf(c, "[%v] is NOT a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
			}
		}

		if command.ExactMatch != "" && (command.ExactMatch == messageText || whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			log.Debugf(c, "%v matched by command.exactMatch", command.Code)
			matchedCommand = &command
			return
		}

		if command.DefaultTitle(whc) == messageText {
			log.Debugf(c, "%v matched by command.GetFullName()", command.Code)
			matchedCommand = &command
			return
			// } else {
			// log.Debugf(c, "command(code=%v).Title(whc): %v", command.ByCode, command.DefaultTitle(whc))
		}
		if command.Matcher != nil && command.Matcher(command, whc) {
			log.Debugf(c, "%v matched by command.matcher()", command.Code)
			matchedCommand = &command
			return
		}

		if !awaitingReplyCommandFound {
			awaitingReplyToPath := botsfwmodels.AwaitingReplyToPath(awaitingReplyTo)
			if awaitingReplyToPath == string(command.Code) || strings.HasSuffix(awaitingReplyToPath, botsfwmodels.AwaitingReplyToPathSeparator+string(command.Code)) {
				awaitingReplyCommand = command
				switch {
				case awaitingReplyToPath == string(command.Code):
					log.Debugf(c, "%v matched by: awaitingReplyToPath == command.ByCode", command.Code)
				case strings.HasSuffix(awaitingReplyToPath, botsfwmodels.AwaitingReplyToPathSeparator+string(command.Code)):
					log.Debugf(c, "%v matched by: strings.HasSuffix(awaitingReplyToPath, AwaitingReplyToPathSeparator + command.ByCode)", command.Code)
				}
				awaitingReplyCommandFound = true
				continue
			}
		}
		// log.Debugf(c, "%v - not matched, matchedCommand: %v", command.ByCode, matchedCommand)
	}
	if awaitingReplyCommandFound {
		matchedCommand = &awaitingReplyCommand
		// log.Debugf(c, "Assign awaitingReplyCommand to matchedCommand: %v", awaitingReplyCommand.ByCode)
	} else {
		matchedCommand = nil
		// log.Debugf(c, "Cleaning up matchedCommand: %v", matchedCommand)
	}
	return
}

// DispatchInlineQuery dispatches inlines query
func (whRouter *webhooksRouter) DispatchInlineQuery(responder WebhookResponder) {
	panic(fmt.Errorf("not implemented, responder: %+v", responder))
}

func changeLocaleIfLangPassed(whc WebhookContext, callbackUrl *url.URL) (m MessageFromBot, err error) {
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
		//		Text: "Unknown language: " + lang,
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
		//		Text: "It is already current language",
		//	})
		//	return
		//}
	}
	return
}

// Dispatch a query to commands
func (whRouter *webhooksRouter) Dispatch(webhookHandler WebhookHandler, responder WebhookResponder, whc WebhookContext) (err error) {
	ctx := whc.Context()
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Criticalf(ctx, "*webhooksRouter.Dispatch() => PANIC: %v", err)
	// 	}
	// }()

	inputType := whc.Input().InputType()

	typeCommands, found := whRouter.commandsByType[inputType]
	if !found {
		log.Debugf(ctx, "No commands found to match by inputType: %v", botinput.GetWebhookInputTypeIdNameString(inputType))
		whc.Input().LogRequest()
		logInputDetails(whc, false)
		return
	}

	var (
		matchedCommand *Command
		commandAction  CommandAction
		m              MessageFromBot
	)

	if len(typeCommands.all) == 0 {
		panic("len(typeCommands.all) == 0")
	}

	var isInlineQuery bool

	switch input := whc.Input().(type) {
	case botinput.WebhookCallbackQuery:
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
				commandAction = func(whc WebhookContext) (MessageFromBot, error) {
					return matchedCommand.CallbackAction(whc, callbackURL)
				}
			}
		}
	case botinput.WebhookInlineQuery:
		isInlineQuery = true
		var queryURL *url.URL
		if matchedCommand, queryURL = matchByQueryOrMatcher(whc, input, typeCommands.byCode, func(command Command) bool {
			return command.InlineQueryAction != nil || command.Action != nil
		}); matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0] // TODO: fallback to default command
		}
		if matchedCommand != nil {
			if matchedCommand.InlineQueryAction == nil {
				commandAction = matchedCommand.Action
			} else {
				commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
					return matchedCommand.InlineQueryAction(whc, input, queryURL)
				}
			}
		}
	case botinput.WebhookChosenInlineResult:
		var queryURL *url.URL

		if matchedCommand, queryURL = matchByQueryOrMatcher(whc, input, typeCommands.byCode, func(command Command) bool {
			return command.ChosenInlineResultAction != nil || command.Action != nil
		}); matchedCommand == nil && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0] // TODO: fallback to default command
		}
		if matchedCommand == nil {
			log.Debugf(ctx, "No command found for WebhookChosenInlineResult")
			return nil
		}
		if matchedCommand.ChosenInlineResultAction == nil {
			commandAction = matchedCommand.Action
		} else {
			commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
				return matchedCommand.ChosenInlineResultAction(whc, input, queryURL)
			}
		}
	case botinput.WebhookTextMessage:
		messageText := input.Text()
		isCommandText := strings.HasPrefix(messageText, "/")
		matchedCommand = whRouter.matchMessageCommands(whc, input, isCommandText, messageText, "", typeCommands.all)
		if matchedCommand != nil {
			if isCommandText && messageText[:len("/start")] == "/start" && matchedCommand.StartAction != nil {
				commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
					return matchedCommand.StartAction(whc, messageText)
				}
			} else if matchedCommand.TextAction == nil {
				commandAction = matchedCommand.Action
			} else {
				commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
					return matchedCommand.TextAction(whc, messageText)
				}
			}
		}
	case botinput.WebhookPreCheckoutQuery:
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
			commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
				return matchedCommand.PreCheckoutQueryAction(whc, input)
			}
		} else if matchedCommand.Action != nil {
			commandAction = matchedCommand.Action
		} else {
			err = fmt.Errorf("matchedCommand(code=%s) has no PreCheckoutQueryAction or Action", matchedCommand.Code)
			return
		}
	case botinput.WebhookSuccessfulPayment:
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
			commandAction = func(whc WebhookContext) (m MessageFromBot, err error) {
				return matchedCommand.SuccessfulPaymentAction(whc, input)
			}
		} else if matchedCommand.Action != nil {
			commandAction = matchedCommand.Action
		} else {
			err = fmt.Errorf("matchedCommand(code=%s) has no SuccessfulPaymentAction or Action", matchedCommand.Code)
			return
		}
	case botinput.WebhookMessage:
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
		if inputType == botinput.WebhookInputUnknown {
			panic("Unknown input type")
		}
		matchedCommand = &typeCommands.all[0]
		commandAction = matchedCommand.Action
	}
	if err != nil {
		err = fmt.Errorf("failed to process input{type=%s} by command{code=%s}: %w",
			botinput.GetWebhookInputTypeIdNameString(whc.Input().InputType()), matchedCommand.Code, err)
		whRouter.processCommandResponseError(whc, matchedCommand, responder, err)
		return
	}

	if matchedCommand == nil {
		log.Debugf(ctx, "whr.matchMessageCommands() => matchedCommand == nil")
		if inputType == botinput.WebhookInputChosenInlineResult {
			return
		}
		whc.Input().LogRequest()
		if m = webhookHandler.HandleUnmatched(whc); m.Text != "" || m.BotMessage != nil {
			whRouter.processCommandResponse(matchedCommand, responder, whc, m, nil)
			return
		}
		if chat := whc.Input().Chat(); chat != nil && chat.IsGroupChat() {
			// m = MessageFromBot{Text: "@" + whc.GetBotCode() + ": " + whc.Translate(MessageTextBotDidNotUnderstandTheCommand), Format: MessageFormatHTML}
			// whr.processCommandResponse(matchedCommand, responder, whc, m, nil)
		} else if !isInlineQuery {
			m = whc.NewMessageByCode(MessageTextBotDidNotUnderstandTheCommand)
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
						if _, sendErr := whc.Responder().SendMessage(ctx, whc.NewMessage("Failed to save botChat data: "+err.Error()), BotAPISendMessageOverHTTPS); sendErr != nil {
							log.Errorf(ctx, "Failed to send error message to user: %v", sendErr)
						}
					}
				}
			}

		}
		whRouter.processCommandResponse(matchedCommand, responder, whc, m, err)
	}
	return
}

func logInputDetails(whc WebhookContext, isKnownType bool) {
	c := whc.Context()
	inputType := whc.Input().InputType()
	input := whc.Input()
	inputTypeIdName := botinput.GetWebhookInputTypeIdNameString(inputType)
	logMessage := fmt.Sprintf("webhooksRouter.Dispatch() => WebhookIputType=%s, %T", inputTypeIdName, input)
	switch inputType {
	case botinput.WebhookInputText:
		textMessage := input.(botinput.WebhookTextMessage)
		logMessage += fmt.Sprintf("message text: [%s]", textMessage.Text())
		if textMessage.IsEdited() { // TODO: Should be in app logic, move out of botsfw
			m := whc.NewMessage("🙇 Sorry, editing messages is not supported. Please send a new message.")
			log.Warningf(c, "TODO: Edited messages are not supported by framework yet. Move check to app.")
			_, err := whc.Responder().SendMessage(c, m, BotAPISendMessageOverResponse)
			if err != nil {
				log.Errorf(c, "failed to send message: %v", err)
			}
			return
		}
	case botinput.WebhookInputContact:
		contact := input.(botinput.WebhookContactMessage)
		contactFirstName := contact.GetFirstName()
		contactBotUserID := contact.GetBotUserID()
		logMessage += fmt.Sprintf("contact number: {UserID: %s, FirstName: %s}", contactBotUserID, contactFirstName)
	case botinput.WebhookInputInlineQuery:
		logMessage += fmt.Sprintf("inline query: [%s]", input.(botinput.WebhookInlineQuery).GetQuery())
	case botinput.WebhookInputCallbackQuery:
		logMessage += fmt.Sprintf("callback data: [%s]", input.(botinput.WebhookCallbackQuery).GetData())
	case botinput.WebhookInputChosenInlineResult:
		chosenResult := input.(botinput.WebhookChosenInlineResult)
		logMessage += fmt.Sprintf("ChosenInlineResult: ResultID=[%s], InlineMessageID=[%s], Query=[%s]", chosenResult.GetResultID(), chosenResult.GetInlineMessageID(), chosenResult.GetQuery())
	case botinput.WebhookInputReferral:
		referralMessage := input.(botinput.WebhookReferralMessage)
		logMessage += fmt.Sprintf("referralMessage: Type=[%s], Source=[%s], Ref=[%s]", referralMessage.Type(), referralMessage.Source(), referralMessage.RefData())
	default:
		logMessage += "Unknown WebhookInputType=" + botinput.GetWebhookInputTypeIdNameString(inputType)
	}
	if isKnownType {
		log.Debugf(c, logMessage)
	} else {
		log.Warningf(c, logMessage)
	}

	m := whc.NewMessage(fmt.Sprintf("Unknown WebhookInputType=%d", inputType)) // TODO: Move out of framework to app?
	_, err := whc.Responder().SendMessage(c, m, BotAPISendMessageOverResponse)
	if err != nil {
		log.Errorf(c, "Failed to send message: %v", err)
	}
}

func (whRouter *webhooksRouter) processCommandResponse(matchedCommand *Command, responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	if err != nil {
		whRouter.processCommandResponseError(whc, matchedCommand, responder, err)
		return
	}

	c := whc.Context()

	responseChannel := m.ResponseChannel
	if responseChannel == "" {
		responseChannel = BotAPISendMessageOverResponse
	}
	if _, err = responder.SendMessage(c, m, responseChannel); err != nil {
		const failedToSendMessageToMessenger = "failed to send a message to messenger"
		errText := err.Error()
		switch {
		case strings.Contains(errText, "message is not modified"): // TODO: This checks are specific to Telegram and should be abstracted or moved to TG related package
			logText := failedToSendMessageToMessenger
			if whc.Input().InputType() == botinput.WebhookInputCallbackQuery {
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
	if matchedCommand != nil {
		path := string(matchedCommand.Code)
		title := matchedCommand.Title
		//if inputType := whc.Input().InputType(); inputType != botinput.WebhookInputCallbackQuery {
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
		//		path = botinput.GetWebhookInputTypeIdNameString(inputType)
		//		title = matchedCommand.Title
		//	}
		//}

		if path != "" {
			platformID := whc.BotPlatform().ID()
			botCode := whc.GetBotCode()
			pageView := analytics.NewPageview(platformID, "bot/"+botCode+"/"+path).SetURL(platformID + "://" + botCode + "/" + path)
			if title != "" {
				pageView = pageView.SetTitle(title)
			}
			pageView.Properties().Set("bot", botCode)
			whAnalytics := whc.Analytics()
			whAnalytics.Enqueue(pageView)
		}
	}
}

func (whRouter *webhooksRouter) processCommandResponseError(whc WebhookContext, matchedCommand *Command, responder WebhookResponder, err error) {
	ctx := whc.Context()
	// log.Errorf() we are logging this in dispatcher
	env := whc.GetBotSettings().Env

	if env == EnvProduction {
		whc.Analytics().Enqueue(analytics.NewErrorMessage(err))
	}
	//inputType := whc.Input().InputType()
	switch inputType := whc.Input().InputType(); inputType {
	case botinput.WebhookInputText, botinput.WebhookInputContact:
		// TODO: Try to get botChat ID from user?
		m := whc.NewMessage(
			whc.Translate(MessageTextOopsSomethingWentWrong) +
				"\n\n" +
				"💢" +
				fmt.Sprintf(" Server error - failed to process message: %v", err),
		)

		if whRouter.errorFooterText != nil {
			args := ErrorFooterArgs{
				BotCode:      whc.GetBotCode(),
				BotProfileID: "", // TODO(help-wanted): implement!
			}
			if footer := whRouter.errorFooterText(ctx, args); footer != "" {
				m.Text += "\n\n" + footer
			}
		}
		if _, respErr := responder.SendMessage(ctx, m, BotAPISendMessageOverResponse); respErr != nil {
			log.Errorf(ctx, "Failed to report to user a server error for command %T: %v", matchedCommand, respErr)
		}
	case botinput.WebhookInputCallbackQuery:
		// TODO: For Telegram call answerInlineQuery to report error to user.
		logus.Errorf(ctx, "Failed to process callback query by command{code=%s}: %v", matchedCommand.Code, inputType)
	default:
		logus.Errorf(ctx, "Failed to process %v input by command{code=%s}: %v", inputType, matchedCommand.Code, inputType)
	}
}
