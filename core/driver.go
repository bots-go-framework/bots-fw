package bots

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"github.com/astec/go-ogle-analytics"
	"strings"
)

// The driver is doing initial request & final response processing
// That includes logging, creating input messages in a general format, sending response
type WebhookDriver interface {
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

type BotDriver struct {
	botHost    BotHost
	appContext BotAppContext
	router     *WebhooksRouter
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

func NewBotDriver(appContext BotAppContext, host BotHost, router *WebhooksRouter) WebhookDriver {
	return BotDriver{appContext: appContext, botHost: host, router: router}
}

func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	logger := d.botHost.Logger(r)
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
			gam, gaErr := ga.NewClientWithHttpClient(d.router.GaTrackingID, botContext.BotHost.GetHttpClient(r))
			if gaErr == nil {
				go func(){
					gaErr = gam.Send(ga.NewException(messageText, true))
				}()
			} else {
				logger.Errorf("Failed to send exception details to GA: %v", gaErr)
			}
			//whc.ReplyByBot(whc.NewMessage(emoji.PANIC_ERROR + " " + messageText))
		}
	}()

	if err != nil {
		logger.Errorf("Failed to create new WebhookContext: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	botCoreStores := webhookHandler.CreateBotCoreStores(d.appContext, r)
	defer func() {
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
					callbackQuery := input.InputCallbackQuery()
					callbackData := callbackQuery.GetData()
					logger.Infof("Input[%v].InputCallbackQuery().GetData(): %v", j, callbackData)
				case WebhookInputInlineQuery:
					logger.Infof("Input[%v].InputInlineQuery().GetQuery(): %v", j, input.InputInlineQuery().GetQuery())
				case WebhookInputChosenInlineResult:
					logger.Infof("Input[%v].InputChosenInlineResult().GetInlineMessageID(): %v", j, input.InputChosenInlineResult().GetInlineMessageID())
				}
				whc := webhookHandler.CreateWebhookContext(d.appContext, r, botContext, input, botCoreStores)
				if whc.GetBotSettings().Mode == Development && !strings.Contains(r.Host, "dev") {
					logger.Warningf("whc.GetBotSettings().Mode == Development && !strings.Contains(r.Host, 'dev')")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if whc.GetBotSettings().Mode == Staging && !strings.Contains(r.Host, "st1") {
					logger.Warningf("whc.GetBotSettings().Mode == Staging && !strings.Contains(r.Host, 'st1')")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
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
