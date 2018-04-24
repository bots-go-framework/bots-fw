package bots

import (
	"fmt"
	//"net/http"
	"net/url"
	"strings"

	"github.com/DebtsTracker/translations/emoji"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"github.com/strongo/gamp"
)

// TypeCommands container for commands
type TypeCommands struct {
	all    []Command
	byCode map[string]Command
}

func newTypeCommands(commandsCount int) *TypeCommands {
	return &TypeCommands{
		byCode: make(map[string]Command, commandsCount),
		all:    make([]Command, 0, commandsCount),
	}
}

func (v *TypeCommands) addCommand(command Command, commandType WebhookInputType) {
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

// WebhooksRouter maps routes to commands
type WebhooksRouter struct {
	commandsByType  map[WebhookInputType]*TypeCommands
	errorFooterText func() string
}

// NewWebhookRouter creates new router
func NewWebhookRouter(commandsByType map[WebhookInputType][]Command, errorFooterText func() string) WebhooksRouter {
	r := WebhooksRouter{
		commandsByType:  make(map[WebhookInputType]*TypeCommands, len(commandsByType)),
		errorFooterText: errorFooterText,
	}

	if commandsByType != nil {
		for commandsType, commands := range commandsByType {
			r.AddCommands(commandsType, commands)
		}
	}

	return r
}

// AddCommands add commands to a router
func (router *WebhooksRouter) AddCommands(commandsType WebhookInputType, commands []Command) {
	typeCommands, ok := router.commandsByType[commandsType]
	if !ok {
		typeCommands = newTypeCommands(len(commands))
		router.commandsByType[commandsType] = typeCommands
	} else if commandsType == WebhookInputInlineQuery {
		panic("Duplicate add of WebhookInputInlineQuery")
	}
	if commandsType == WebhookInputInlineQuery && len(commands) > 1 {
		panic("commandsType == WebhookInputInlineQuery && len(commands) > 1")
	}
	for _, command := range commands {
		typeCommands.addCommand(command, commandsType)
	}
	if commandsType == WebhookInputInlineQuery && len(typeCommands.all) > 1 {
		panic(fmt.Sprintf("commandsType == WebhookInputInlineQuery && len(typeCommands) > 1: %v", typeCommands.all[0]))
	}
}

// RegisterCommands is registering commands with router
func (router *WebhooksRouter) RegisterCommands(commands []Command) {
	addCommand := func(t WebhookInputType, command Command) {
		typeCommands, ok := router.commandsByType[t]
		if !ok {
			typeCommands = newTypeCommands(0)
			router.commandsByType[t] = typeCommands
		}
		typeCommands.addCommand(command, t)
	}
	for _, command := range commands {
		if len(command.InputTypes) == 0 {
			if command.Action != nil {
				addCommand(WebhookInputText, command)
			}
			if command.CallbackAction != nil {
				addCommand(WebhookInputCallbackQuery, command)
			}
		} else {
			for _, t := range command.InputTypes {
				addCommand(t, command)
			}
		}
	}
}

func matchCallbackCommands(whc WebhookContext, input WebhookCallbackQuery, typeCommands *TypeCommands) (matchedCommand *Command, callbackURL *url.URL, err error) {
	if len(typeCommands.all) > 0 {
		callbackData := input.GetData()
		callbackURL, err = url.Parse(callbackData)
		if err != nil {
			log.Errorf(whc.Context(), "Failed to parse callback data to URL: %v", err.Error())
		} else {
			callbackPath := callbackURL.Path
			if command, ok := typeCommands.byCode[callbackPath]; ok {
				return &command, callbackURL, nil
			}
		}
		if err == nil && matchedCommand == nil {
			err = fmt.Errorf("No commands matchet to callback: [%v]", callbackData)
			whc.LogRequest()
		}
	} else {
		panic("len(typeCommands.all) == 0")
	}
	return nil, callbackURL, err
}

func (router *WebhooksRouter) matchMessageCommands(whc WebhookContext, input WebhookMessage, parentPath string, commands []Command) (matchedCommand *Command) {
	var (
		messageText, messageTextLowerCase string
		awaitingReplyCommand              Command
	)

	c := whc.Context()

	//if parentPath == "" {
	//	log.Debugf(c, "matchMessageCommands()")
	//}

	if textMessage, ok := input.(WebhookTextMessage); ok {
		messageText = textMessage.Text()
		messageTextLowerCase = strings.ToLower(messageText)
	}

	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	//log.Debugf(c, "awaitingReplyTo: %v", awaitingReplyTo)

	var awaitingReplyCommandFound bool

	{
		commandText := messageTextLowerCase
		if strings.HasPrefix(commandText, "/") && strings.Contains(commandText, "@") {
			commandText = commandText[:strings.Index(commandText, "@")]
		}
		for _, command := range commands {
			for _, commandName := range command.Commands {
				if commandName == commandText || strings.HasPrefix(messageTextLowerCase, commandName+" ") {
					log.Debugf(c, "command(code=%v) matched by command.commands", command.Code)
					matchedCommand = &command
					return
				}
			}
		}
	}

	for _, command := range commands {
		if !awaitingReplyCommandFound && awaitingReplyTo != "" {
			awaitingReplyPrefix := strings.TrimLeft(parentPath+AWAITING_REPLY_TO_PATH_SEPARATOR+command.Code, AWAITING_REPLY_TO_PATH_SEPARATOR)

			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				//log.Debugf(c, "[%v] is a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
				//log.Debugf(c, "awaitingReplyCommand: %v", command.ByCode)
				if matchedCommand = router.matchMessageCommands(whc, input, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
					log.Debugf(c, "%v matched by command.replies", command.Code)
					awaitingReplyCommand = *matchedCommand
					awaitingReplyCommandFound = true
					continue
				}
			} else {
				//log.Debugf(c, "[%v] is NOT a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
			}
		}

		if command.ExactMatch != "" && (command.ExactMatch == messageText || whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			log.Debugf(c, "%v matched by command.exactMatch", command.Code)
			matchedCommand = &command
			return
		}

		if command.DefaultTitle(whc) == messageText {
			log.Debugf(c, "%v matched by command.FullName()", command.Code)
			matchedCommand = &command
			return
			// } else {
			//log.Debugf(c, "command(code=%v).Title(whc): %v", command.ByCode, command.DefaultTitle(whc))
		}
		if command.Matcher != nil && command.Matcher(command, whc) {
			log.Debugf(c, "%v matched by command.matcher()", command.Code)
			matchedCommand = &command
			return
		}

		if !awaitingReplyCommandFound {
			awaitingReplyToPath := AwaitingReplyToPath(awaitingReplyTo)
			if awaitingReplyToPath == command.Code || strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR+command.Code) {
				awaitingReplyCommand = command
				switch {
				case awaitingReplyToPath == command.Code:
					log.Debugf(c, "%v matched by: awaitingReplyToPath == command.ByCode", command.Code)
				case strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR+command.Code):
					log.Debugf(c, "%v matched by: strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR + command.ByCode)", command.Code)
				}
				awaitingReplyCommandFound = true
				continue
			}
		}
		//log.Debugf(c, "%v - not matched, matchedCommand: %v", command.ByCode, matchedCommand)
	}
	if awaitingReplyCommandFound {
		matchedCommand = &awaitingReplyCommand
		//log.Debugf(c, "Assign awaitingReplyCommand to matchedCommand: %v", awaitingReplyCommand.ByCode)
	} else {
		matchedCommand = nil
		//log.Debugf(c, "Cleaning up matchedCommand: %v", matchedCommand)
	}
	return
}

// DispatchInlineQuery dispatches inlines query
func (router *WebhooksRouter) DispatchInlineQuery(responder WebhookResponder) {

}

// Dispatch query to commands
func (router *WebhooksRouter) Dispatch(responder WebhookResponder, whc WebhookContext) {
	c := whc.Context()
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Criticalf(c, "*WebhooksRouter.Dispatch() => PANIC: %v", err)
	//	}
	//}()

	inputType := whc.InputType()

	typeCommands, found := router.commandsByType[inputType]
	if !found {
		log.Debugf(c, "No commands found to match by inputType: %v", WebhookInputTypeNames[inputType])
		whc.LogRequest()
		logInputDetails(whc, false)
		return
	}

	var (
		matchedCommand *Command
		commandAction  CommandAction
		err            error
		m              MessageFromBot
	)
	input := whc.Input()
	switch input.(type) {
	case WebhookCallbackQuery:
		var callbackURL *url.URL
		matchedCommand, callbackURL, err = matchCallbackCommands(whc, input.(WebhookCallbackQuery), typeCommands)
		if err == nil && matchedCommand != nil {
			if matchedCommand.Code == "" {
				err = fmt.Errorf("matchedCommand(%T: %v).ByCode is empty string", matchedCommand, matchedCommand)
			} else if matchedCommand.CallbackAction == nil {
				err = fmt.Errorf("matchedCommand(%T: %v).CallbackAction == nil", matchedCommand, matchedCommand.Code)
			} else {
				log.Debugf(c, "matchCallbackCommands() => matchedCommand: %T(code=%v)", matchedCommand, matchedCommand.Code)
				commandAction = func(whc WebhookContext) (MessageFromBot, error) {
					return matchedCommand.CallbackAction(whc, callbackURL)
				}
			}
		}
	case WebhookMessage:
		inputType := input.InputType()
		if inputType == WebhookInputNewChatMembers && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0]
		}
		if matchedCommand == nil {
			matchedCommand = router.matchMessageCommands(whc, input.(WebhookMessage), "", typeCommands.all)
			if matchedCommand != nil {
				log.Debugf(c, "router.matchMessageCommands() => matchedCommand.Code: %v", matchedCommand.Code)
			}
		}
		if matchedCommand != nil {
			commandAction = matchedCommand.Action
		}
	default:
		if inputType == WebhookInputUnknown {
			panic("Unknown input type")
		}
		matchedCommand = &typeCommands.all[0]
		commandAction = matchedCommand.Action
	}
	if err != nil {
		router.processCommandResponse(matchedCommand, responder, whc, m, err)
		return
	}

	if matchedCommand == nil {
		whc.LogRequest()
		log.Debugf(c, "router.matchMessageCommands() => matchedCommand == nil")
		if whc.Chat().IsGroupChat() {
			//m = MessageFromBot{Text: "@" + whc.GetBotCode() + ": " + whc.Translate(MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND), Format: MessageFormatHTML}
			//router.processCommandResponse(matchedCommand, responder, whc, m, nil)
		} else {
			m = whc.NewMessageByCode(MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND)
			chatEntity := whc.ChatEntity()
			if chatEntity != nil && chatEntity.GetAwaitingReplyTo() != "" {
				m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %v</i>", chatEntity.GetAwaitingReplyTo())
			}
			log.Debugf(c, "No command found for the message: %v", input)
			router.processCommandResponse(matchedCommand, responder, whc, m, nil)
		}
	} else {
		if matchedCommand.Code == "" {
			log.Debugf(c, "Matched to: %v", matchedCommand)
		} else {
			log.Debugf(c, "Matched to: %v", matchedCommand.Code) //runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
		}
		var err error
		if commandAction == nil {
			err = errors.New("No action for matched command")
		} else {
			m, err = commandAction(whc)
		}
		router.processCommandResponse(matchedCommand, responder, whc, m, err)
	}
}

func logInputDetails(whc WebhookContext, isKnownType bool) {
	c := whc.Context()
	inputType := whc.InputType()
	input := whc.Input()
	logMessage := fmt.Sprintf("WebhooksRouter.Dispatch() => inputType: %v=%v, %T", inputType, WebhookInputTypeNames[inputType], input)
	if !isKnownType {
		logMessage += fmt.Sprintf(" => no commands to match for input type=%v", WebhookInputTypeNames[inputType])
	}
	switch inputType {
	case WebhookInputText:
		textMessage := input.(WebhookTextMessage)
		logMessage += fmt.Sprintf("message text: [%v]", textMessage.Text())
		if textMessage.IsEdited() { // TODO: Should be in app logic, move out of core
			m := whc.NewMessage("🙇 Sorry, editing messages is not supported. Please send a new message.")
			log.Warningf(c, "TODO: Edited messages are not supported by framework yet. Move check to app.")
			whc.Responder().SendMessage(c, m, BotApiSendMessageOverResponse)
			return
		}
	case WebhookInputContact:
		logMessage += fmt.Sprintf("contact number: [%v]", input.(WebhookContactMessage))
	case WebhookInputInlineQuery:
		logMessage += fmt.Sprintf("inline query: [%v]", input.(WebhookInlineQuery).GetQuery())
	case WebhookInputCallbackQuery:
		logMessage += fmt.Sprintf("callback data: [%v]", input.(WebhookCallbackQuery).GetData())
	case WebhookInputChosenInlineResult:
		chosenResult := input.(WebhookChosenInlineResult)
		logMessage += fmt.Sprintf("ChosenInlineResult: ResultID=[%v], InlineMessageID=[%v], Query=[%v]", chosenResult.GetResultID(), chosenResult.GetInlineMessageID(), chosenResult.GetQuery())
	case WebhookInputReferral:
		referralMessage := input.(WebhookReferralMessage)
		logMessage += fmt.Sprintf("referralMessage: Type=[%v], Source=[%v], Ref=[%v]", referralMessage.Type(), referralMessage.Source(), referralMessage.RefData())
	}
	if isKnownType {
		log.Debugf(c, logMessage)
	} else {
		log.Warningf(c, logMessage)
	}

	m := whc.NewMessage("Sorry, unknown input type") // TODO: Move out of framework to app.
	whc.Responder().SendMessage(c, m, BotApiSendMessageOverResponse)

	return
}

func (router *WebhooksRouter) processCommandResponse(matchedCommand *Command, responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {

	c := whc.Context()
	gaMeasurement := whc.GaMeasurement()
	//gam.GeographicalOverride()

	env := whc.GetBotSettings().Env
	inputType := whc.InputType()
	if err == nil {
		if _, err = responder.SendMessage(c, m, BotApiSendMessageOverHTTPS); err != nil {
			const failedToSendMessageToMessenger = "failed to send a message to messenger"
			if strings.Contains(err.Error(), "message is not modified") { // TODO: This check is specific to Telegram and should be abstracted
				logText := failedToSendMessageToMessenger
				if inputType == WebhookInputCallbackQuery {
					logText += "(can be duplicate callback)"
				}
				log.Warningf(c, errors.WithMessage(err, logText).Error()) // TODO: Think how to get rid of warning on duplicate callbacks when users clicks multiple times
				err = nil
			} else {
				log.Errorf(c, errors.WithMessage(err, failedToSendMessageToMessenger).Error()) //TODO: Decide how do we handle this
			}
		}
		if matchedCommand != nil {
			if gaMeasurement != nil {

				gaHostName := fmt.Sprintf("%v.debtstracker.io", strings.ToLower(whc.BotPlatform().Id()))
				pathPrefix := "bot/"
				var pageview gamp.Pageview
				var chatEntity BotChat
				if inputType != WebhookInputCallbackQuery {
					chatEntity = whc.ChatEntity()
				}
				if inputType != WebhookInputCallbackQuery && chatEntity != nil {
					path := chatEntity.GetAwaitingReplyTo()
					if path == "" {
						path = matchedCommand.Code
					} else if pathURL, err := url.Parse(path); err == nil {
						path = pathURL.Path
					}
					pageview = gamp.NewPageviewWithDocumentHost(gaHostName, pathPrefix+path, matchedCommand.Title)
				} else {
					pageview = gamp.NewPageviewWithDocumentHost(gaHostName, pathPrefix+WebhookInputTypeNames[inputType], matchedCommand.Title)
				}

				pageview.Common = whc.GaCommon()
				if err := gaMeasurement.Queue(pageview); err != nil {
					log.Warningf(c, "Failed to send page view to GA: %v", err)
				}
			}
		}
	} else {
		log.Errorf(c, err.Error())
		if env == strongo.EnvProduction && gaMeasurement != nil {
			exceptionMessage := gamp.NewException(err.Error(), false)
			exceptionMessage.Common = whc.GaCommon()
			err = gaMeasurement.Queue(exceptionMessage)
			if err != nil {
				log.Warningf(c, "Failed to send page view to GA: %v", err)
			}
		}
		if inputType == WebhookInputText || inputType == WebhookInputContact {
			// Todo: Try to get chat ID from user?
			m := whc.NewMessage(
				whc.Translate(MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG) +
					"\n\n" +
					emoji.ERROR_ICON +
					fmt.Sprintf(" Server error - failed to process message: %v", err),
			)

			if router.errorFooterText != nil {
				if footer := router.errorFooterText(); footer != "" {
					m.Text += "\n\n" + footer
				}
			}

			if _, respErr := responder.SendMessage(c, m, BotApiSendMessageOverResponse); respErr != nil {
				log.Errorf(c, "Failed to report to user a server error: %v", respErr)
			}
		}
	}
}
