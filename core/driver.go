package bots

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/emoji"
	"fmt"
	"github.com/strongo/measurement-protocol"
	"net/http"
	"runtime/debug"
	"strings"
)

// The driver is doing initial request & final response processing
// That includes logging, creating input messages in a general format, sending response
type WebhookDriver interface {
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

type BotDriver struct {
	GaTrackingID string
	botHost      BotHost
	appContext   BotAppContext
	router       *WebhooksRouter
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

func NewBotDriver(gaTrackingID string, appContext BotAppContext, host BotHost, router *WebhooksRouter) WebhookDriver {
	return BotDriver{GaTrackingID: gaTrackingID, appContext: appContext, botHost: host, router: router}
}

func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	logger := d.botHost.Logger(r)
	c := d.botHost.Context(r) // TODO: It's wrong to have dependency on appengine here
	logger.Infof(c, "HandleWebhook() => webhookHandler: %T", webhookHandler)

	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(r)

	if err != nil {
		if _, ok := err.(AuthFailedError); ok {
			logger.Warningf(c, "Auth failed: %v", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			logger.Errorf(c, "Failed to call webhookHandler.GetBotContextAndInputs(r): %v", err)
			//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	logger.Infof(c, "Got %v entries", len(entriesWithInputs))

	var whc WebhookContext
	gaMeasurement := measurement.NewBufferedSender([]string{d.GaTrackingID}, true, botContext.BotHost.GetHttpClient(r))

	defer func() {
		logger.Debugf(c, "driver.deferred(recover) - checking for panic & flush GA")
		if recovered := recover(); recovered != nil {
			messageText := fmt.Sprintf("Server error (panic): %v", recovered)
			logger.Criticalf(c, "Panic recovered: %s\n%s", messageText, debug.Stack())

			gaMessage := measurement.NewException(messageText, true)

			if whc != nil {
				gaMessage.Common = whc.GaCommon()
			} else {
				gaMessage.Common.ClientID = "c7ea15eb-3333-4d47-a002-9d1a14996371"
				gaMessage.Common.DataSource = "bot"
			}

			if err := gaMeasurement.Queue(gaMessage); err != nil {
				logger.Errorf(c, "Failed to queue exception details for GA: %v", err)
			} else {
				logger.Debugf(c, "Exception details queued for GA.")
			}
			logger.Debugf(c, "Flushing gaMeasurement (with exeception, len(queue): %v)...", gaMeasurement.QueueDepth())
			if err = gaMeasurement.Flush(); err != nil {
				logger.Errorf(c, "Failed to send exception details to GA: %v", err)
			} else {
				logger.Debugf(c, "Exception details sent to GA.")
			}
			if whc != nil {
				if whc.BotChatID() != nil {
					whc.Responder().SendMessage(c, whc.NewMessage(emoji.ERROR_ICON+" "+messageText), BotApiSendMessageOverResponse)
				}
			}
		} else {
			logger.Debugf(c, "Flushing gaMeasurement (len(queue): %v)...", gaMeasurement.QueueDepth())
			gaMeasurement.Flush()
		}
	}()

	if err != nil {
		logger.Errorf(c, "Failed to create new WebhookContext: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	botCoreStores := webhookHandler.CreateBotCoreStores(d.appContext, r)
	defer func() {
		if whc != nil {
			logger.Debugf(c, "Closing BotChatStore...")
			chatEntity := whc.ChatEntity()
			if chatEntity != nil && chatEntity.GetPreferredLanguage() == "" {
				chatEntity.SetPreferredLanguage(whc.Locale().Code5)
			}
			if err := botCoreStores.BotChatStore.Close(); err != nil {
				logger.Errorf(c, "Failed to close BotChatStore: %v", err)
			}
		}
	}()

	for i, entryWithInputs := range entriesWithInputs {
		logger.Infof(c, "Entry[%v]: %v, %v inputs", i, entryWithInputs.Entry.GetID(), len(entryWithInputs.Inputs))
		for j, input := range entryWithInputs.Inputs {
			inputType := input.InputType()
			switch inputType {
			case WebhookInputMessage, WebhookInputInlineQuery, WebhookInputCallbackQuery, WebhookInputChosenInlineResult:
				switch inputType {
				case WebhookInputMessage:
					logger.Infof(c, "Input[%v].Message().Text(): %v", j, input.InputMessage().Text())
				case WebhookInputCallbackQuery:
					callbackQuery := input.InputCallbackQuery()
					callbackData := callbackQuery.GetData()
					logger.Infof(c, "Input[%v].InputCallbackQuery().GetData(): %v", j, callbackData)
				case WebhookInputInlineQuery:
					logger.Infof(c, "Input[%v].InputInlineQuery().GetQuery(): %v", j, input.InputInlineQuery().GetQuery())
				case WebhookInputChosenInlineResult:
					logger.Infof(c, "Input[%v].InputChosenInlineResult().GetInlineMessageID(): %v", j, input.InputChosenInlineResult().GetInlineMessageID())
				}
				whc = webhookHandler.CreateWebhookContext(d.appContext, r, botContext, input, botCoreStores, gaMeasurement.New(botContext.BotSettings.Mode != Production))
				if whc.GetBotSettings().Mode == Development && !strings.Contains(r.Host, "dev") {
					logger.Warningf(c, "whc.GetBotSettings().Mode == Development && !strings.Contains(r.Host, 'dev')")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if whc.GetBotSettings().Mode == Staging && !strings.Contains(r.Host, "st1") {
					logger.Warningf(c, "whc.GetBotSettings().Mode == Staging && !strings.Contains(r.Host, 'st1')")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				responder := webhookHandler.GetResponder(w, whc)
				d.router.Dispatch(responder, whc)
			case WebhookInputUnknown:
				logger.Warningf(c, "Unknown input[%v] type", j)
			default:
				logger.Warningf(c, "Unhandled input[%v] type: %v=%v", j, inputType, WebhookInputTypeNames[inputType])
			}
		}
	}
}
