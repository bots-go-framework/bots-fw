package bots

import (
	"fmt"
	//"net/http"
	"strings"
)

type WebhooksRouter struct {
	commandsByType map[WebhookInputType][]Command
	commandsByCode map[string]Command
}

func NewWebhookRouter(commandsByType map[WebhookInputType][]Command) *WebhooksRouter {
	r := &WebhooksRouter{
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

	for _, command := range commands {

		if len(command.Replies) > 0 || parentPath != "" {
			var awaitingReplyPrefix string
			if parentPath == "" {
				awaitingReplyPrefix = command.Code
			} else {
				awaitingReplyPrefix = strings.Join([]string{parentPath, command.Code}, AWAITING_REPLY_TO_PATH_SEPARATOR)
			}
			awaitingReplyTo := whc.ChatEntity().GetAwaitingReplyTo()
			logger.Debugf("awaitingReplyPrefix: %v; awaitingReplyTo: %v", awaitingReplyPrefix, awaitingReplyTo)
			if strings.HasPrefix(awaitingReplyTo, awaitingReplyPrefix) {
				awaitingReplyCommand = command
				logger.Debugf("awaitingReplyCommand: %v", awaitingReplyCommand.Code)
				if matchedCommand = r.matchCommands(whc, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
					logger.Debugf("%v matched my command.replies", command.Code)
					return
				}
			}
		}

		if command.ExactMatch != "" && (command.ExactMatch == messageText || whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			logger.Debugf("%v matched my command.exactMatch", command.Code)
			matchedCommand = &command
			return
		}

		if command.DefaultTitle(whc) == messageText {
			logger.Debugf("%v matched my command.Title()", command.Code)
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
		logger.Debugf("%v - not matched, matchedCommand: %v", command.Code, matchedCommand)
	}
	if awaitingReplyCommand.Code != "" {
		matchedCommand = &awaitingReplyCommand
		logger.Debugf("Assign awaitingReplyCommand to matchedCommand: %v", awaitingReplyCommand)
	} else {
		matchedCommand = nil
		logger.Debugf("Cleanin up matchedCommand: %v", matchedCommand)
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
			processCommandResponse(responder, whc, m, nil)
		} else {
			logger.Infof("Matched to: %v", matchedCommand.Code) //runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
			m, err := matchedCommand.Action(whc)
			processCommandResponse(responder, whc, m, err)
			return
		}
	} else {
		logger.Infof("No commands found byt input type %v=%v", inputType, WebhookInputTypeNames[inputType])
	}
}

func processCommandResponse(responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	logger := whc.GetLogger()
	if err == nil {
		logger.Infof("Bot response message: %v", m)
		err = responder.SendMessage(m)
		if err != nil {
			logger.Errorf("Failed to send message to Telegram\n\tError: %v\n\tMessage text: %v", err, m.Text) //TODO: Decide how do we handle it
		}
	} else {
		logger.Errorf(err.Error())
		if whc.InputType() == WebhookInputMessage { // Todo: Try to get chat ID from user?
			err = responder.SendMessage(whc.NewMessage(whc.Translate(MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG) + "\n\n" + fmt.Sprintf("\xF0\x9F\x9A\xA8 Server error - failed to process message: %v", err)))
			if err != nil {
				logger.Errorf("Failed to report to user a server error: %v", err)
			}
		}
	}
}
