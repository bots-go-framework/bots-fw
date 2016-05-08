package bots

import (
	"fmt"
	//"net/http"
	"strings"
)

type WebhooksRouter struct {
	commands       []Command
	commandsByCode map[string]Command
}

func NewWebhookRouter(commands []Command) *WebhooksRouter {
	r := &WebhooksRouter{
		commands: commands,
	}
	r.verifyCommands()
	return r
}
func (r *WebhooksRouter) verifyCommands() {
	r.commandsByCode = make(map[string]Command, len(r.commands))
	for _, command := range r.commands {
		if command.Code == "" {
			panic(fmt.Sprintf("Command %v is missing required property Code", command))
		}
		if _, ok := r.commandsByCode[command.Code]; ok {
			panic(fmt.Sprintf("Command with code '%v' defined multiple times", command.Code))
		}
		r.commandsByCode[command.Code] = command
	}
}

func (r *WebhooksRouter) matchCommands(whc WebhookContext, parentPath string, commands []Command) (matchedCommand *Command) {
	messageText := whc.MessageText()
	messageTextLowerCase := strings.ToLower(messageText)

	var awaitingReplyCommand Command

	log := whc.GetLogger()

	for _, command := range commands {
		if len(command.Replies) > 0 || parentPath != "" {
			var awaitingReplyPrefix string
			if parentPath == "" {
				awaitingReplyPrefix = command.Code
			} else {
				awaitingReplyPrefix = strings.Join([]string{parentPath, command.Code}, AWAITING_REPLY_TO_PATH_SEPARATOR)
			}
			log.Debugf("awaitingReplyPrefix: %v", awaitingReplyPrefix)
			if strings.HasPrefix(whc.ChatEntity().GetAwaitingReplyTo(), awaitingReplyPrefix) {
				awaitingReplyCommand = command
				log.Debugf("awaitingReplyCommand: %v", awaitingReplyCommand.Code)
				if matchedCommand = r.matchCommands(whc, awaitingReplyPrefix, command.Replies); matchedCommand != nil {
					log.Debugf("%v matched my command.replies", command.Code)
					return
				}
			}
		}
		if command.ExactMatch == messageText || (command.ExactMatch != "" && whc.TranslateNoWarning(command.ExactMatch) == messageText) {
			log.Debugf("%v matched my command.exactMatch", command.Code)
			matchedCommand = &command
			return
		}
		if command.DefaultTitle(whc) == messageText {
			log.Debugf("%v matched my command.Title()", command.Code)
			matchedCommand = &command
			return
		} else {
			log.Debugf("command(code=%v).Title(whc): %v", command.Code, command.DefaultTitle(whc))
		}
		for _, commandName := range command.Commands {
			if messageTextLowerCase == commandName || strings.HasPrefix(messageTextLowerCase, commandName+" ") {
				log.Debugf("%v matched my command.commands", command.Code)
				matchedCommand = &command
				return
			}
		}
		if command.Matcher != nil && command.Matcher(command, whc) {
			log.Debugf("%v matched my command.matcher()", command.Code)
			matchedCommand = &command
			return
		}
		log.Debugf("%v - not matched, matchedCommand: %v", command.Code, matchedCommand)
	}
	if awaitingReplyCommand.Code != "" {
		matchedCommand = &awaitingReplyCommand
		log.Debugf("Assign awaitingReplyCommand to matchedCommand: %v", awaitingReplyCommand)
	} else {
		matchedCommand = nil
		log.Debugf("Cleanin up matchedCommand: %v", matchedCommand)
	}

	log.Debugf("matchedCommand: %v", matchedCommand)
	return
}

func (r *WebhooksRouter) Dispatch(responder WebhookResponder, whc WebhookContext) {
	log := whc.GetLogger()
	chatEntity := whc.ChatEntity()
	log.Debugf("message text: [%v]", whc.InputMessage().Text())
	//log.Debugf("AwaitingReplyTo: [%v]", chatEntity.GetAwaitingReplyTo())

	matchedCommand := r.matchCommands(whc, "", r.commands)

	if matchedCommand == nil {
		m := MessageFromBot{Text: whc.Translate(MESSAGE_TEXT_I_DID_NOT_UNDERSTAND_THE_COMMAND), Format: MessageFormatHTML}
		if chatEntity.GetAwaitingReplyTo() != "" {
			m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %v</i>", chatEntity.GetAwaitingReplyTo())
		}
		log.Infof("No command found for the message: %v", whc.MessageText())
		processCommandResponse(responder, whc, m, nil)
	} else {
		log.Infof("Matched to: %v", matchedCommand.Code) //runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
		m, err := matchedCommand.Action(whc)
		processCommandResponse(responder, whc, m, err)
		return
	}
}

func processCommandResponse(responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	log := whc.GetLogger()
	if err == nil {
		log.Infof("Bot response message: %v", m)
		//_, err := whc.BotApi(whc.Context()).Send(m)
		//chattable := tgbotapi.Chattable(m)
		err = responder.SendMessage(m)
		if err != nil {
			log.Errorf("Failed to send message to Telegram\n\tError: %v\n\tMessage text: %v", err, m.Text) //TODO: Decide how do we handle it
		}
	} else {
		log.Errorf(err.Error())
		err = responder.SendMessage(whc.NewMessage(whc.Translate(MESSAGE_TEXT_OOPS_SOMETHING_WENT_WRONG) + "\n\n" + fmt.Sprintf("\xF0\x9F\x9A\xA8 Server error - failed to process message: %v", err)))
		if err != nil {
			log.Errorf("Failed to report to user a server error: %v", err)
		}
	}
}
