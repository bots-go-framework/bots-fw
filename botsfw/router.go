package botsfw

import (
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/strongo/gamp"
	"net/url"
	"strings"
	"time"
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
//
//goland:noinspection GoUnusedExportedFunction
func NewWebhookRouter(commandsByType map[WebhookInputType][]Command, errorFooterText func() string) WebhooksRouter {
	r := WebhooksRouter{
		commandsByType:  make(map[WebhookInputType]*TypeCommands, len(commandsByType)),
		errorFooterText: errorFooterText,
	}

	for commandsType, commands := range commandsByType {
		r.AddCommands(commandsType, commands)
	}

	return r
}

func (whRouter *WebhooksRouter) CommandsCount() int {
	var count int
	for _, v := range whRouter.commandsByType {
		count += len(v.all)
	}
	return count
}

// AddCommands add commands to a router
func (whRouter *WebhooksRouter) AddCommands(commandsType WebhookInputType, commands []Command) {
	typeCommands, ok := whRouter.commandsByType[commandsType]
	if !ok {
		typeCommands = newTypeCommands(len(commands))
		whRouter.commandsByType[commandsType] = typeCommands
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
func (whRouter *WebhooksRouter) RegisterCommands(commands []Command) {
	addCommand := func(t WebhookInputType, command Command) {
		typeCommands, ok := whRouter.commandsByType[t]
		if !ok {
			typeCommands = newTypeCommands(0)
			whRouter.commandsByType[t] = typeCommands
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
			callbackAdded := false
			for _, t := range command.InputTypes {
				addCommand(t, command)
				if t == WebhookInputCallbackQuery {
					callbackAdded = true
				}
			}
			if command.CallbackAction != nil && !callbackAdded {
				addCommand(WebhookInputCallbackQuery, command)
			}
		}
	}
}

var ErrNoCommandsMatched = errors.New("no commands matched")

func matchCallbackCommands(whc WebhookContext, input WebhookCallbackQuery, typeCommands *TypeCommands) (matchedCommand *Command, callbackURL *url.URL, err error) {
	if len(typeCommands.all) > 0 {
		callbackData := input.GetData()
		callbackURL, err = url.Parse(callbackData)
		if err != nil {
			log.Errorf(whc.Context(), "Failed to parse callback data to URL: %v", err.Error())
		} else {
			for _, c := range typeCommands.all {
				if c.Matcher != nil {
					if c.Matcher(c, whc) {
						return &c, callbackURL, nil
					}
				}
			}
			callbackPath := callbackURL.Path
			if command, ok := typeCommands.byCode[callbackPath]; ok {
				return &command, callbackURL, nil
			}
		}
		//if matchedCommand == nil {
		log.Errorf(whc.Context(), fmt.Errorf("%w: %s", ErrNoCommandsMatched, fmt.Sprintf("callbackData=[%v]", callbackData)).Error())
		whc.LogRequest()
		//}
	} else {
		panic("len(typeCommands.all) == 0")
	}
	return nil, callbackURL, err
}

func (whRouter *WebhooksRouter) matchMessageCommands(whc WebhookContext, input WebhookMessage, isCommandText bool, messageText, parentPath string, commands []Command) (matchedCommand *Command) {
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
			awaitingReplyPrefix := strings.TrimLeft(parentPath+botsfwmodels.AwaitingReplyToPathSeparator+command.Code, botsfwmodels.AwaitingReplyToPathSeparator)

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
			if awaitingReplyToPath == command.Code || strings.HasSuffix(awaitingReplyToPath, botsfwmodels.AwaitingReplyToPathSeparator+command.Code) {
				awaitingReplyCommand = command
				switch {
				case awaitingReplyToPath == command.Code:
					log.Debugf(c, "%v matched by: awaitingReplyToPath == command.ByCode", command.Code)
				case strings.HasSuffix(awaitingReplyToPath, botsfwmodels.AwaitingReplyToPathSeparator+command.Code):
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
func (whRouter *WebhooksRouter) DispatchInlineQuery(responder WebhookResponder) {
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

// Dispatch query to commands
func (whRouter *WebhooksRouter) Dispatch(webhookHandler WebhookHandler, responder WebhookResponder, whc WebhookContext) {
	c := whc.Context()
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Criticalf(c, "*WebhooksRouter.Dispatch() => PANIC: %v", err)
	// 	}
	// }()

	inputType := whc.InputType()

	typeCommands, found := whRouter.commandsByType[inputType]
	if !found {
		log.Debugf(c, "No commands found to match by inputType: %v", GetWebhookInputTypeIdNameString(inputType))
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
	var isCommandText bool
	switch input := input.(type) {
	case WebhookCallbackQuery:
		var callbackURL *url.URL
		matchedCommand, callbackURL, err = matchCallbackCommands(whc, input, typeCommands)
		if err == nil && matchedCommand != nil {
			if matchedCommand.Code == "" {
				err = fmt.Errorf("matchedCommand(%T: %v).ByCode is empty string", matchedCommand, matchedCommand)
			} else if matchedCommand.CallbackAction == nil {
				err = fmt.Errorf("matchedCommand(%T: %v).CallbackAction == nil", matchedCommand, matchedCommand.Code)
			} else {
				log.Debugf(c, "matchCallbackCommands() => matchedCommand: %T(code=%v)", matchedCommand, matchedCommand.Code)
				if m, err = changeLocaleIfLangPassed(whc, callbackURL); err != nil || m.Text != "" {
					return
				}
				commandAction = func(whc WebhookContext) (MessageFromBot, error) {
					return matchedCommand.CallbackAction(whc, callbackURL)
				}
			}
		}
	case WebhookMessage:
		if inputType == WebhookInputNewChatMembers && len(typeCommands.all) == 1 {
			matchedCommand = &typeCommands.all[0]
		}
		if matchedCommand == nil {
			var messageText string
			if textMessage, ok := input.(WebhookTextMessage); ok {
				messageText = textMessage.Text()
				isCommandText = strings.HasPrefix(messageText, "/")
			}
			matchedCommand = whRouter.matchMessageCommands(whc, input, isCommandText, messageText, "", typeCommands.all)
			if matchedCommand != nil {
				log.Debugf(c, "whr.matchMessageCommands() => matchedCommand.Code: %v", matchedCommand.Code)
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
		whRouter.processCommandResponseError(whc, matchedCommand, responder, err)
		return
	}

	if matchedCommand == nil {
		whc.LogRequest()
		log.Debugf(c, "whr.matchMessageCommands() => matchedCommand == nil")
		if m = webhookHandler.HandleUnmatched(whc); m.Text != "" || m.BotMessage != nil {
			whRouter.processCommandResponse(matchedCommand, responder, whc, m, nil)
			return
		}
		if chat := whc.Chat(); chat != nil && chat.IsGroupChat() {
			// m = MessageFromBot{Text: "@" + whc.GetBotCode() + ": " + whc.Translate(MessageTextBotDidNotUnderstandTheCommand), Format: MessageFormatHTML}
			// whr.processCommandResponse(matchedCommand, responder, whc, m, nil)
		} else {
			m = whc.NewMessageByCode(MessageTextBotDidNotUnderstandTheCommand)
			chatEntity := whc.ChatData()
			if chatEntity != nil {
				if awaitingReplyTo := chatEntity.GetAwaitingReplyTo(); awaitingReplyTo != "" {
					m.Text += fmt.Sprintf("\n\n<i>AwaitingReplyTo: %v</i>", awaitingReplyTo)
				}
			}
			log.Debugf(c, "No command found for the message: %v", input)
			whRouter.processCommandResponse(matchedCommand, responder, whc, m, nil)
		}
	} else { // matchedCommand != nil
		if matchedCommand.Code == "" {
			log.Debugf(c, "Matched to: %+v", matchedCommand)
		} else {
			log.Debugf(c, "Matched to: %v", matchedCommand.Code) // runtime.FuncForPC(reflect.ValueOf(command.Action).Pointer()).Name()
		}
		var err error
		if commandAction == nil {
			err = errors.New("No action for matched command")
		} else {
			m, err = commandAction(whc)
			// awaitingReplyToAfter := chatData.GetAwaitingReplyTo()
			// if isCommandText && awaitingReplyToAfter == awaitingReplyToBefore { // TODO: Looks dangerous? Should be commands be responsible?
			// 	log.Debugf(c, "Auto-resetting AwaitingReplyTo when not changed after processing and isCommandText=true")
			// 	chatData.SetAwaitingReplyTo("")
			// }
		}
		if err == nil {
			if chatData := whc.ChatData(); chatData != nil {
				now := time.Now()
				chatData.SetDtLastInteraction(now)
				if chatData.IsChanged() {
					chatData.SetUpdatedTime(now)
				}
				if err = whc.SaveBotChat(c); err != nil {
					log.Errorf(c, "Failed to save botChat data: %v", err)
					if _, sendErr := whc.Responder().SendMessage(c, whc.NewMessage("Failed to save botChat data: "+err.Error()), BotAPISendMessageOverHTTPS); sendErr != nil {
						log.Errorf(c, "Failed to send error message to user: %v", sendErr)
					}
				}
			}

		}
		whRouter.processCommandResponse(matchedCommand, responder, whc, m, err)
	}
}

func logInputDetails(whc WebhookContext, isKnownType bool) {
	c := whc.Context()
	inputType := whc.InputType()
	input := whc.Input()
	inputTypeIdName := GetWebhookInputTypeIdNameString(inputType)
	logMessage := fmt.Sprintf("WebhooksRouter.Dispatch() => WebhookIputType=%s, %T", inputTypeIdName, input)
	switch inputType {
	case WebhookInputText:
		textMessage := input.(WebhookTextMessage)
		logMessage += fmt.Sprintf("message text: [%v]", textMessage.Text())
		if textMessage.IsEdited() { // TODO: Should be in app logic, move out of botsfw
			m := whc.NewMessage("🙇 Sorry, editing messages is not supported. Please send a new message.")
			log.Warningf(c, "TODO: Edited messages are not supported by framework yet. Move check to app.")
			_, err := whc.Responder().SendMessage(c, m, BotAPISendMessageOverResponse)
			if err != nil {
				log.Errorf(c, "failed to send message: %v", err)
			}
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
	default:
		logMessage += "Unknown WebhookInputType=" + GetWebhookInputTypeIdNameString(inputType)
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

func (whRouter *WebhooksRouter) processCommandResponse(matchedCommand *Command, responder WebhookResponder, whc WebhookContext, m MessageFromBot, err error) {
	if err != nil {
		whRouter.processCommandResponseError(whc, matchedCommand, responder, err)
		return
	}

	c := whc.Context()
	ga := whc.GA()
	// gam.GeographicalOverride()

	if _, err = responder.SendMessage(c, m, BotAPISendMessageOverResponse); err != nil {
		const failedToSendMessageToMessenger = "failed to send a message to messenger"
		errText := err.Error()
		switch {
		case strings.Contains(errText, "message is not modified"): // TODO: This checks are specific to Telegram and should be abstracted or moved to TG related package
			logText := failedToSendMessageToMessenger
			if whc.InputType() == WebhookInputCallbackQuery {
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
	if matchedCommand != nil && ga != nil {
		gaHostName := fmt.Sprintf("%v.debtstracker.io", strings.ToLower(whc.BotPlatform().ID()))
		pathPrefix := "bot/"
		var pageview *gamp.Pageview
		if inputType := whc.InputType(); inputType != WebhookInputCallbackQuery {
			chatData := whc.ChatData()
			if chatData != nil {
				path := chatData.GetAwaitingReplyTo()
				if path == "" {
					path = matchedCommand.Code
				} else if pathURL, err := url.Parse(path); err == nil {
					path = pathURL.Path
				}
				pageview = gamp.NewPageviewWithDocumentHost(gaHostName, pathPrefix+path, matchedCommand.Title)
			} else {
				pageview = gamp.NewPageviewWithDocumentHost(gaHostName, pathPrefix+GetWebhookInputTypeIdNameString(inputType), matchedCommand.Title)
			}
		}

		pageview.Common = ga.GaCommon()
		if err := ga.Queue(pageview); err != nil {
			if strings.Contains(err.Error(), "no tracking ID") {
				log.Debugf(c, "process command response: failed to send page view to GA: %v", err)
			} else {
				log.Warningf(c, "proess command response: failed to send page view to GA: %v", err)
			}

		}
	}
}

func (whRouter *WebhooksRouter) processCommandResponseError(whc WebhookContext, matchedCommand *Command, responder WebhookResponder, err error) {
	c := whc.Context()
	log.Errorf(c, err.Error())
	env := whc.GetBotSettings().Env
	ga := whc.GA()
	if env == EnvProduction && ga != nil {
		exceptionMessage := gamp.NewException(err.Error(), false)
		exceptionMessage.Common = ga.GaCommon()
		err = ga.Queue(exceptionMessage)
		if err != nil {
			if strings.Contains(err.Error(), "no tracking ID") {
				log.Debugf(c, "processCommandResponseError: failed to send page view to GA: %v", err)
			} else {
				log.Warningf(c, "processCommandResponseError: failed to send page view to GA: %v", err)
			}
		}
	}
	inputType := whc.InputType()
	if inputType == WebhookInputText || inputType == WebhookInputContact {
		// TODO: Try to get botChat ID from user?
		m := whc.NewMessage(
			whc.Translate(MessageTextOopsSomethingWentWrong) +
				"\n\n" +
				"💢" +
				fmt.Sprintf(" Server error - failed to process message: %v", err),
		)

		if whRouter.errorFooterText != nil {
			if footer := whRouter.errorFooterText(); footer != "" {
				m.Text += "\n\n" + footer
			}
		}

		if _, respErr := responder.SendMessage(c, m, BotAPISendMessageOverResponse); respErr != nil {
			log.Errorf(c, "Failed to report to user a server error for command %T: %v", matchedCommand, respErr)
		}
	}
}
