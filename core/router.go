package bots

import (
	"fmt"
	//"net/http"
	"strings"
	"bitbucket.com/debtstracker/gae_app/debtstracker/emoji"
	"github.com/strongo/measurement-protocol"
	"net/url"
	"github.com/pkg/errors"
)

type TypeCommands struct {
	all    []Command
	byCode map[string]Command
}

type WebhooksRouter struct {
	commandsByType map[WebhookInputType]TypeCommands
}

func NewWebhookRouter(commandsByType map[WebhookInputType][]Command) *WebhooksRouter {
	r := &WebhooksRouter{commandsByType: make(map[WebhookInputType]TypeCommands, len(commandsByType))}
	for commandType, commands := range commandsByType {
		commandsCount := len(commands)
		typeCommands := TypeCommands{
			byCode: make(map[string]Command, commandsCount),
			all: make([]Command, commandsCount, commandsCount),
		}
		for i, command := range commands {
			if command.Code == "" {
				panic(fmt.Sprintf("Command %v is missing required property Code", command))
			}
			if _, ok := typeCommands.byCode[command.Code]; ok {
				panic(fmt.Sprintf("Command with code '%v' defined multiple times", command.Code))
			}
			typeCommands.all[i] = command
			typeCommands.byCode[command.Code] = command
		}
		r.commandsByType[commandType] = typeCommands
	}
	return r
}

func matchCallbackCommands (whc WebhookContext, typeCommands TypeCommands) (matchedCommand *Command, callbackUrl *url.URL, err error) {
	if len(typeCommands.all) > 0 {
		callbackData := whc.InputCallbackQuery().GetData()
		callbackUrl, err = url.Parse(callbackData)
		if err != nil {
			whc.Logger().Errorf("Failed to parse callback data to URL: %v", err.Error())
		} else {
			callbackPath := callbackUrl.Path
			if command, ok := typeCommands.byCode[callbackPath]; ok {
				return &command, callbackUrl, nil
			}
		}
	}
	return nil, callbackUrl, err
}

func (r *WebhooksRouter) matchFirstCommand(commands []Command) (matchedCommand *Command) {
	matchedCommand = &commands[0]
	return
}

func (r *WebhooksRouter) matchMessageCommands(whc WebhookContext, parentPath string, commands []Command) (matchedCommand *Command) {
	var awaitingReplyCommand Command

	logger := whc.Logger()

	messageText := whc.MessageText()
	messageTextLowerCase := strings.ToLower(messageText)

	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	//logger.Debugf("awaitingReplyTo: %v", awaitingReplyTo)

	var awaitingReplyCommandFound bool

	for _, command := range commands {
		for _, commandName := range command.Commands {
			if messageTextLowerCase == commandName || strings.HasPrefix(messageTextLowerCase, commandName + " ") {
				logger.Debugf("command(code=%v) matched my command.commands", command.Code)
				matchedCommand = &command
				return
			}
		}
	}

	for _, command := range commands {
		if !awaitingReplyCommandFound && awaitingReplyTo != "" {
			awaitingReplyPrefix := strings.TrimLeft(parentPath + AWAITING_REPLY_TO_PATH_SEPARATOR + command.Code, AWAITING_REPLY_TO_PATH_SEPARATOR)

			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				//logger.Debugf("[%v] is a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
				//logger.Debugf("awaitingReplyCommand: %v", command.Code)
				if matchedCommand = r.matchMessageCommands(whc, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
					logger.Debugf("%v matched my command.replies", command.Code)
					awaitingReplyCommand = *matchedCommand
					awaitingReplyCommandFound = true
					continue
				}
			} else {
				logger.Debugf("[%v] is NOT a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
			}
		}

		if command.ExactMatch != "" && (command.ExactMatch == messageText || whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			logger.Debugf("%v matched my command.exactMatch", command.Code)
			matchedCommand = &command
			return
		}

		if command.DefaultTitle(whc) == messageText {
			logger.Debugf("%v matched my command.FullName()", command.Code)
			matchedCommand = &command
			return
		} else {
			logger.Debugf("command(code=%v).Title(whc): %v", command.Code, command.DefaultTitle(whc))
		}
		if command.Matcher != nil && command.Matcher(command, whc) {
			logger.Debugf("%v matched my command.matcher()", command.Code)
			matchedCommand = &command
			return
		}

		if !awaitingReplyCommandFound {
			awaitingReplyToPath := AwaitingReplyToPath(awaitingReplyTo)
			if awaitingReplyToPath == command.Code || strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR + command.Code) {
				awaitingReplyCommand = command
				switch {
				case awaitingReplyToPath == command.Code:
					logger.Debugf("%v matched by: awaitingReplyToPath == command.Code", command.Code)
				case strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR + command.Code):
					logger.Debugf("%v matched by: strings.HasSuffix(awaitingReplyToPath, AWAITING_REPLY_TO_PATH_SEPARATOR + command.Code)", command.Code)
				}
				awaitingReplyCommandFound = true
				continue
			}
		}
		logger.Debugf("%v - not matched, matchedCommand: %v", command.Code, matchedCommand)
	}
	if awaitingReplyCommandFound {
		matchedCommand = &awaitingReplyCommand
		logger.Debugf("Assign awaitingReplyCommand to matchedCommand: %v", awaitingReplyCommand.Code)
	} else {
		matchedCommand = nil
		logger.Debugf("Cleaning up matchedCommand: %v", matchedCommand)
	}

	logger.Debugf("matchedCommand: %v", matchedCommand)
	return
}

func (r *WebhooksRouter) DispatchInlineQuery(responder WebhookResponder) {

}

func (r *WebhooksRouter) Dispatch(responder WebhookResponder, whc WebhookContext) {
	logger := whc.Logger()
	inputType := whc.InputType()

	logMessage := fmt.Sprintf("WebhooksRouter.Dispatch(): inputType: %v=%v, ", inputType, WebhookInputTypeNames[inputType])
	switch inputType {
	case WebhookInputMessage:
		logMessage += fmt.Sprintf("message text: [%v]", whc.InputMessage().Text())
	case WebhookInputInlineQuery:
		logMessage += fmt.Sprintf("inline query: [%v]", whc.InputInlineQuery().GetQuery())
	case WebhookInputCallbackQuery:
		logMessage += fmt.Sprintf("callback data: [%v]", whc.InputCallbackQuery().GetData())
	case WebhookInputChosenInlineResult:
		chosenResult := whc.InputChosenInlineResult()
		logMessage += fmt.Sprintf("ChosenInlineResult: ResultID=[%v], InlineMessageID=[%v], Query=[%v]", chosenResult.GetResultID(), chosenResult.GetInlineMessageID(), chosenResult.GetQuery())
	}


	if typeCommands, found := r.commandsByType[inputType]; !found {
		logMessage += "no commands to match"
		logger.Warningf(logMessage)
		err := errors.New(logMessage)
		var m MessageFromBot
		processCommandResponse(nil, responder, whc, m, err)
	} else {
		logMessage += fmt.Sprintf(", len(commandsToMatch): %v", len(typeCommands.all))
		logger.Debugf(logMessage)

		var matchedCommand *Command
		var commandAction CommandAction
		var err error
		var m MessageFromBot
		switch inputType {
		case WebhookInputCallbackQuery:
			var callbackUrl *url.URL
			matchedCommand, callbackUrl, err = matchCallbackCommands(whc, typeCommands)
			if err == nil {
				if matchedCommand.Code == "" {
					err = errors.New(fmt.Sprintf("matchedCommand(%T: %v).Code is empty string", matchedCommand, matchedCommand))
				} else {
					if matchedCommand.CallbackAction == nil {
						err = errors.New(fmt.Sprintf("matchedCommand(%v).CallbackAction == nil", matchedCommand.Code))
					} else {
						commandAction = func(whc WebhookContext) (MessageFromBot, error) {
							return matchedCommand.CallbackAction(whc, callbackUrl)
						}
					}
				}
			}
		case WebhookInputMessage:
			matchedCommand = r.matchMessageCommands(whc, "", typeCommands.all)
			if matchedCommand != nil {
				commandAction = matchedCommand.Action
			}
		case WebhookInputUnknown:
			panic("Unknown input type")
		default:
			matchedCommand = r.matchFirstCommand(typeCommands.all)
		}
		if err != nil {
			processCommandResponse(matchedCommand, responder, whc, m, err)
			return
		}

		if matchedCommand == nil {
			m = MessageFromBot{Text: whc.Translate(MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND), Format: MessageFormatHTML}
			chatEntity := whc.ChatEntity()
			if chatEntity != nil && chatEntity.GetAwaitingReplyTo() != "" {
				m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %v</i>", chatEntity.GetAwaitingReplyTo())
			}
			logger.Infof("No command found for the message: %v", whc.MessageText())
			processCommandResponse(matchedCommand, responder, whc, m, nil)
		} else {
			logger.Infof("Matched to: %v", matchedCommand.Code) //runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
			m, err := commandAction(whc)
			processCommandResponse(matchedCommand, responder, whc, m, err)
		}
	}
}

func processCommandResponse(matchedCommand *Command, responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	logger := whc.Logger()
	gaMeasurement := whc.GaMeasurement()
	//gam.GeographicalOverride()

	if err == nil {
		logger.Infof("processCommandResponse(): Bot response message: %v", m)
		if _, err = responder.SendMessage(m, BotApiSendMessageOverResponse); err != nil {
			logger.Errorf("Failed to send message to Telegram\n\tError: %v\n\tMessage text: %v", err, m.Text) //TODO: Decide how do we handle it
		}
		if matchedCommand != nil {
			if gaMeasurement != nil {
				chatEntity := whc.ChatEntity()
				gaHostName := fmt.Sprintf("%v.debtstracker.io", strings.ToLower(whc.BotPlatform().Id()))
				pathPrefix := "bot/"
				var pageview measurement.Pageview
				if chatEntity != nil {
					path := chatEntity.GetAwaitingReplyTo()
					if path == "" {
						path = matchedCommand.Code
					}
					pageview = measurement.NewPageviewWithDocumentHost(gaHostName, pathPrefix + path, matchedCommand.Title)
				} else {
					pageview = measurement.NewPageviewWithDocumentHost(gaHostName, pathPrefix + WebhookInputTypeNames[whc.InputType()], matchedCommand.Title)
				}
				pageview.Common = whc.GaCommon()

				go func() {
					err := gaMeasurement.Queue(pageview)
					if err != nil {
						logger.Warningf("Failed to send page view to GA: %v", err)
					}
				}()
			}
		}
	} else {
		logger.Errorf(err.Error())
		if gaMeasurement != nil {
			exceptionMessage := measurement.NewException(err.Error(), false)
			exceptionMessage.Common = whc.GaCommon()
			go func(){
				err = gaMeasurement.Queue(exceptionMessage)
				if err != nil {
					logger.Warningf("Failed to send page view to GA: %v", err)
				}
			}()
		}
		if whc.InputType() == WebhookInputMessage {
			// Todo: Try to get chat ID from user?
			_, respErr := responder.SendMessage(whc.NewMessage(whc.Translate(MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG) + "\n\n" + emoji.ERROR_ICON + fmt.Sprintf(" Server error - failed to process message: %v", err)), BotApiSendMessageOverResponse)
			if respErr != nil {
				logger.Errorf("Failed to report to user a server error: %v", respErr)
			}
		}
	}
}
