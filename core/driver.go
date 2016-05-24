package bots

import (
	"net/http"
	"fmt"
	"runtime/debug"
)

// The driver is doing initial request & final response processing
// That includes logging, creating input messages in a general format, sending response
type WebhookDriver interface {
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

type BotDriver struct {
	botHost BotHost
	appContext AppContext
	router  *WebhooksRouter
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

func NewBotDriver(appContext AppContext, host BotHost, router *WebhooksRouter) WebhookDriver {
	return BotDriver{appContext: appContext, botHost: host, router: router}
}

func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	logger := d.botHost.GetLogger(r)
	logger.Infof("HandleWebhook() => webhookHandler: %T", webhookHandler)

	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(r)

	if err != nil {
		if _, ok := err.(AuthFailedError); ok {
			logger.Warningf("Auth failed: %v", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			logger.Errorf("Failed to call webhookHandler.GetBotContext(r): %v", err)
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	logger.Infof("Got %v entries", len(entriesWithInputs))

	defer func() {
		if recovered := recover(); recovered != nil {
			messageText := fmt.Sprintf("Server error (panic): %v", recovered)
			logger.Criticalf("Panic recovered: %s\n%s", messageText, debug.Stack())
			//whc.ReplyByBot(whc.NewMessage("\xF0\x9F\x9A\xA8 " + messageText))
		}
	}()

	if err != nil {
		logger.Errorf("Failed to create new WebhookContext: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	botCoreStores := webhookHandler.CreateBotCoreStores(d.appContext, r)
	defer func(){
		logger.Debugf("Closing BotChatStore...")
		if err := botCoreStores.BotChatStore.Close(); err != nil {
			logger.Errorf("Failed to close BotChatStore: %v", err)
		}
	}()

	for i, entryWithInputs := range entriesWithInputs {
		logger.Infof("Entry[%v]: %v, %v inputs", i, entryWithInputs.Entry.GetID(), len(entryWithInputs.Inputs))
		for j, input := range entryWithInputs.Inputs {
			inputType := input.InputType()
			switch inputType {
			case WebhookInputMessage, WebhookInputInlineQuery, WebhookInputCallbackQuery, WebhookInputChosenInlineResult:
				switch inputType {
				case WebhookInputMessage:
					logger.Infof("Input[%v].Message().Text(): %v", j, input.InputMessage().Text())
				case WebhookInputCallbackQuery:
					logger.Infof("Input[%v].InputCallbackQuery().GetData(): %v", j, input.InputCallbackQuery().GetData())
				case WebhookInputInlineQuery:
					logger.Infof("Input[%v].InputInlineQuery().GetQuery(): %v", j, input.InputInlineQuery().GetQuery())
				case WebhookInputChosenInlineResult:
					logger.Infof("Input[%v].InputChosenInlineResult().GetInlineMessageID(): %v", j, input.InputChosenInlineResult().GetInlineMessageID())
				}
				whc := webhookHandler.CreateWebhookContext(d.appContext, r, botContext, input, botCoreStores)
				responder := webhookHandler.GetResponder(w, whc)
				d.router.Dispatch(responder, whc)
			case WebhookInputUnknown:
				logger.Warningf("Unknown input[%v] type", j)
			default:
				logger.Warningf("Unhandled input[%v] type: %v=%v", j, inputType, WebhookInputTypeNames[inputType])
			}
		}
	}
}
