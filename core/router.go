package bots

import (
	"fmt"
	//"net/http"
	"strings"
	"github.com/astec/go-ogle-analytics"
	"strconv"
)

type WebhooksRouter struct {
	GaTrackingID string
	commandsByType map[WebhookInputType][]Command
	commandsByCode map[string]Command
}

func NewWebhookRouter(gaTrackingID string, commandsByType map[WebhookInputType][]Command) *WebhooksRouter {
	r := &WebhooksRouter{
		GaTrackingID: gaTrackingID,
		commandsByType: commandsByType,
		commandsByCode: make(map[string]Command, len(commandsByType)),
	}
	for _, commands := range commandsByType {
		for _, command := range commands {
			if command.Code == "" {
				panic(fmt.Sprintf("Command %v is missing required property Code", command))
			}
			if _, ok := r.commandsByCode[command.Code]; ok {
				panic(fmt.Sprintf("Command with code '%v' defined multiple times", command.Code))
			}
			r.commandsByCode[command.Code] = command
		}
	}
	return r
}

func (r *WebhooksRouter) matchCommands(whc WebhookContext, parentPath string, commands []Command) (matchedCommand *Command) {
	inputType := whc.InputType()

	var awaitingReplyCommand Command

	logger := whc.GetLogger()

	if inputType != WebhookInputMessage && inputType != WebhookInputUnknown && len(commands) == 1 {
		command := commands[0]
		if command.InputType == inputType { //TODO: it should refactored - we match 1st by type for now
			logger.Debugf("%v matched my command.InputType", command.Code)
			matchedCommand = &command
			return
		} else {
			logger.Warningf("inputType: %v, commandInputType: %v, commandCode: %v", WebhookInputTypeNames[inputType], WebhookInputTypeNames[command.InputType], command.Code)
		}
	}

	messageText := whc.MessageText()
	messageTextLowerCase := strings.ToLower(messageText)

	awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
	//logger.Debugf("awaitingReplyTo: %v", awaitingReplyTo)

	var awaitingReplyCommandFound bool

	for _, command := range commands {

		if !awaitingReplyCommandFound && awaitingReplyTo != "" {
			awaitingReplyPrefix := strings.TrimLeft(parentPath + AWAITING_REPLY_TO_PATH_SEPARATOR + command.Code, AWAITING_REPLY_TO_PATH_SEPARATOR)

			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				//logger.Debugf("[%v] is a prefix for [%v]", awaitingReplyPrefix, awaitingReplyTo)
				//logger.Debugf("awaitingReplyCommand: %v", command.Code)
				if matchedCommand = r.matchCommands(whc, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
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
		for _, commandName := range command.Commands {
			if messageTextLowerCase == commandName || strings.HasPrefix(messageTextLowerCase, commandName+" ") {
				logger.Debugf("%v matched my command.commands", command.Code)
				matchedCommand = &command
				return
			}
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
	logger := whc.GetLogger()
	inputType := whc.InputType()
	switch inputType {
	case WebhookInputMessage:
		logger.Debugf("message text: [%v]", whc.InputMessage().Text())
	case WebhookInputInlineQuery:
		logger.Debugf("inline query: [%v]", whc.InputInlineQuery().GetQuery())
	case WebhookInputCallbackQuery:
		logger.Debugf("callback data: [%v]", whc.InputCallbackQuery().GetData())
	case WebhookInputChosenInlineResult:
		chosenResult := whc.InputChosenInlineResult()
		logger.Debugf("ChosenInlineResult: ResultID=[%v], InlineMessageID=[%v], Query=[%v]", chosenResult.GetResultID(), chosenResult.GetInlineMessageID(), chosenResult.GetQuery())
	}

	if commands, found := r.commandsByType[inputType]; found {
		matchedCommand := r.matchCommands(whc, "", commands)

		if matchedCommand == nil {
			m := MessageFromBot{Text: whc.Translate(MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND), Format: MessageFormatHTML}
			chatEntity := whc.ChatEntity()
			if chatEntity != nil && chatEntity.GetAwaitingReplyTo() != "" {
				m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %v</i>", chatEntity.GetAwaitingReplyTo())
			}
			logger.Infof("No command found for the message: %v", whc.MessageText())
			processCommandResponse(r.GaTrackingID, matchedCommand, responder, whc, m, nil)
		} else {
			logger.Infof("Matched to: %v", matchedCommand.Code) //runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
			m, err := matchedCommand.Action(whc)
			processCommandResponse(r.GaTrackingID, matchedCommand, responder, whc, m, err)
		}
	} else {
		logger.Infof("No commands found byt input type %v=%v", inputType, WebhookInputTypeNames[inputType])
	}
}

func processCommandResponse(gaTrackingID string, matchedCommand *Command, responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	logger := whc.GetLogger()
	gam, gaErr := ga.NewClientWithHttpClient(gaTrackingID, whc.GetHttpClient())
	//gam.GeographicalOverride()
	gam.ClientID(strconv.FormatInt(whc.AppUserIntID(), 10))
	if gaErr != nil {
		logger.Errorf("Failed to create client with TrackingID: [%v]", gaTrackingID)
		panic(err)
	}
	if err == nil {
		logger.Infof("processCommandResponse(): Bot response message: %v", m)
		if err = responder.SendMessage(m, BotApiSendMessageOverResponse); err != nil {
			logger.Errorf("Failed to send message to Telegram\n\tError: %v\n\tMessage text: %v", err, m.Text) //TODO: Decide how do we handle it
		}
		if matchedCommand != nil {
			if gam != nil {
				chatEntity := whc.ChatEntity()
				gaHostName := fmt.Sprintf("%v.debtstracker.io", strings.ToLower(whc.BotPlatform().Id()))
				pathPrefix := "bot/"
				if chatEntity != nil {
					path := chatEntity.GetAwaitingReplyTo()
					if path == "" {
						path = matchedCommand.Code
					}
					go func() {
						gaErr = gam.Send(ga.NewPageview(gaHostName, pathPrefix + path, matchedCommand.Title))
						if gaErr != nil {
							logger.Warningf("Failed to send page view to GA: %v", gaErr)
						}
					}()
				} else {
					go func() {
						pageview := ga.NewPageview(gaHostName, pathPrefix + WebhookInputTypeNames[whc.InputType()], matchedCommand.Title)
						gaErr = gam.Send(pageview)
						if gaErr != nil {
							logger.Warningf("Failed to send page view to GA: %v", gaErr)
						}
					}()
				}
			}
		}
	} else {
		logger.Errorf(err.Error())
		if whc.InputType() == WebhookInputMessage {
			// Todo: Try to get chat ID from user?
			respErr := responder.SendMessage(whc.NewMessage(whc.Translate(MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG) + "\n\n" + fmt.Sprintf("\xF0\x9F\x9A\xA8 Server error - failed to process message: %v", err)), BotApiSendMessageOverResponse)
			if respErr != nil {
				logger.Errorf("Failed to report to user a server error: %v", respErr)
			}
		}
		if gam != nil {
			exceptionMessage := ga.NewException(err.Error(), false)
			go func(){
				gaErr = gam.Send(exceptionMessage)
				if err != nil {
					logger.Warningf("Failed to send page view to GA: %v", gaErr)
				}
			}()
		}
	}
}
