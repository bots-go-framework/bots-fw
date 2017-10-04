package bots

import (
	"fmt"
	//"net/http"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/pkg/errors"
	"github.com/strongo/measurement-protocol"
	"net/url"
	"strings"
	"github.com/strongo/app/log"
	"github.com/strongo/app"
)

type TypeCommands struct {
	all    []Command
	byCode map[string]Command
}

func newTypeCommands(commandsCount int) *TypeCommands {
	return &TypeCommands{
		byCode: make(map[string]Command, commandsCount),
		all:    make([]Command, commandsCount, commandsCount),
	}
}

func (v *TypeCommands) addCommand(i int, command Command, commandType WebhookInputType) {
	if command.Code == "" {
		panic(fmt.Sprintf("Command %v is missing required property ByCode", command))
	}
	if i < 0 {
		v.all = append(v.all, command)
	} else {
		v.all[i] = command
	}
	if _, ok := v.byCode[command.Code]; !ok {
		v.byCode[command.Code] = command
	} else {
		panic(fmt.Sprintf("Duplicate command code for %v : %v", commandType, command.Code))
	}
}

type WebhooksRouter struct {
	commandsByType  map[WebhookInputType]*TypeCommands
	errorFooterText func() string
}

func NewWebhookRouter(commandsByType map[WebhookInputType][]Command, errorFooterText func() string) *WebhooksRouter {
	r := &WebhooksRouter{
		commandsByType:  make(map[WebhookInputType]*TypeCommands, len(commandsByType)),
		errorFooterText: errorFooterText,
	}

	if commandsByType != nil {
		for commandType, commands := range commandsByType {
			typeCommands := newTypeCommands(len(commands))
			for i, command := range commands {
				typeCommands.addCommand(i, command, commandType)
			}
			r.commandsByType[commandType] = typeCommands
		}
	}

	return r
}

func (router *WebhooksRouter) RegisterCommands(commands []Command) {
	addCommand := func(t WebhookInputType, command Command) {
		typeCommands, ok := router.commandsByType[t]
		if !ok {
			typeCommands = newTypeCommands(0)
			router.commandsByType[t] = typeCommands
		}
		typeCommands.addCommand(-1, command, t)
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

func matchCallbackCommands(whc WebhookContext, input WebhookCallbackQuery, typeCommands *TypeCommands) (matchedCommand *Command, callbackUrl *url.URL, err error) {
	if len(typeCommands.all) > 0 {
		callbackData := input.GetData()
		callbackUrl, err = url.Parse(callbackData)
		if err != nil {
			log.Errorf(whc.Context(), "Failed to parse callback data to URL: %v", err.Error())
		} else {
			callbackPath := callbackUrl.Path
			if command, ok := typeCommands.byCode[callbackPath]; ok {
				return &command, callbackUrl, nil
			}
		}
		if err == nil && matchedCommand == nil {
			err = errors.New(fmt.Sprintf("No commands matchet to callback: [%v]", callbackData))
			whc.LogRequest()
		}
	} else {
		panic("len(typeCommands.all) == 0")
	}
	return nil, callbackUrl, err
}

func (router *WebhooksRouter) matchMessageCommands(whc WebhookContext, input WebhookMessage, parentPath string, commands []Command) (matchedCommand *Command) {
	var (
		messageText, messageTextLowerCase string
		awaitingReplyCommand              Command
	)

	c := whc.Context()

	if parentPath == "" {
		log.Debugf(c, "matchMessageCommands()")
	}

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
		} else {
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

func (router *WebhooksRouter) DispatchInlineQuery(responder WebhookResponder) {

}

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
		whc.LogRequest()
		logInputDetails(whc, false)
		return
	}
	log.Debugf(c, "Found %d commands to match by inputType: %v", len(typeCommands.all), inputType)
	var (
		matchedCommand *Command
		commandAction  CommandAction
		err            error
		m              MessageFromBot
	)
	input := whc.Input()
	switch input.(type) {
	case WebhookCallbackQuery:
		var callbackUrl *url.URL
		matchedCommand, callbackUrl, err = matchCallbackCommands(whc, input.(WebhookCallbackQuery), typeCommands)
		if err == nil && matchedCommand != nil {
			if matchedCommand.Code == "" {
				err = errors.New(fmt.Sprintf("matchedCommand(%T: %v).ByCode is empty string", matchedCommand, matchedCommand))
			} else if matchedCommand.CallbackAction == nil {
				err = errors.New(fmt.Sprintf("matchedCommand(%T: %v).CallbackAction == nil", matchedCommand, matchedCommand.Code))
			} else {
				log.Debugf(c, "matchCallbackCommands() => matchedCommand: %T(code=%v)", matchedCommand, matchedCommand.Code)
				commandAction = func(whc WebhookContext) (MessageFromBot, error) {
					return matchedCommand.CallbackAction(whc, callbackUrl)
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
				log.Debugf(c, "router.matchMessageCommands() => matchedCommand: %v", matchedCommand)
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
			const FAILED_TO_SEND_MESSAGE_TO_MESSENGER = "failed to send a message to messenger"
			if strings.Contains(err.Error(), "message is not modified") { // TODO: This check is specific to Telegram and should be abstracted
				logText := FAILED_TO_SEND_MESSAGE_TO_MESSENGER
				if inputType == WebhookInputCallbackQuery {
					logText += "(can be duplicate callback)"
				}
				log.Warningf(c, errors.WithMessage(err, logText).Error()) // TODO: Think how to get rid of warning on duplicate callbacks when users clicks multiple times
				err = nil
			} else {
				log.Errorf(c, errors.WithMessage(err, FAILED_TO_SEND_MESSAGE_TO_MESSENGER).Error()) //TODO: Decide how do we handle this
			}
		}
		if matchedCommand != nil {
			if gaMeasurement != nil {

				gaHostName := fmt.Sprintf("%v.debtstracker.io", strings.ToLower(whc.BotPlatform().Id()))
				pathPrefix := "bot/"
				var pageview measurement.Pageview
				var chatEntity BotChat
				if inputType != WebhookInputCallbackQuery {
					chatEntity = whc.ChatEntity()
				}
				if inputType != WebhookInputCallbackQuery && chatEntity != nil {
					path := chatEntity.GetAwaitingReplyTo()
					if path == "" {
						path = matchedCommand.Code
					} else if pathUrl, err := url.Parse(path); err == nil {
						path = pathUrl.Path
					}
					pageview = measurement.NewPageviewWithDocumentHost(gaHostName, pathPrefix+path, matchedCommand.Title)
				} else {
					pageview = measurement.NewPageviewWithDocumentHost(gaHostName, pathPrefix+WebhookInputTypeNames[inputType], matchedCommand.Title)
				}

				pageview.Common = whc.GaCommon()
				err := gaMeasurement.Queue(pageview)
				if err != nil {
					log.Warningf(c, "Failed to send page view to GA: %v", err)
				}
			}
		}
	} else {
		log.Errorf(c, err.Error())
		if env == strongo.EnvProduction && gaMeasurement != nil {
			exceptionMessage := measurement.NewException(err.Error(), false)
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
