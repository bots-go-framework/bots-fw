package bots

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/emoji"
	"fmt"
	"github.com/strongo/measurement-protocol"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// The driver is doing initial request & final response processing
// That includes logging, creating input messages in a general format, sending response
type WebhookDriver interface {
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

type BotDriver struct {
	GaSettings GaSettings
	botHost    BotHost
	appContext BotAppContext
	router     *WebhooksRouter
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

type GaSettings struct {
	TrackingID string
	Enabled func(r *http.Request) bool
}

func NewBotDriver(gaSettings GaSettings, appContext BotAppContext, host BotHost, router *WebhooksRouter) WebhookDriver {
	return BotDriver{GaSettings: gaSettings, appContext: appContext, botHost: host, router: router}
}

func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	started := time.Now()
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

	var (
		whc WebhookContext
		gaMeasurement *measurement.BufferedSender
	)
	{  // Initiate Google Analytics Measurement API client
		var sendStats bool
		if d.GaSettings.Enabled == nil {
			sendStats = botContext.BotSettings.Mode == Production
			logger.Debugf(c, "d.GaSettings.Enabled == nil, botContext.BotSettings.Mode: %v, sendStats: %v", botContext.BotSettings.Mode, sendStats)
		} else {
			sendStats = d.GaSettings.Enabled(r)
			logger.Debugf(c, "d.GaSettings.Enabled != nil, sendStats: %v", sendStats)
		}
		if sendStats {
			trackingID := d.GaSettings.TrackingID
			botHost := botContext.BotHost
			gaMeasurement = measurement.NewBufferedSender([]string{trackingID}, true, botHost.GetHttpClient(r))
		} else {
			gaMeasurement = measurement.NewDiscardingBufferedSender()
		}
	}

	defer func() {
		logger.Debugf(c, "driver.deferred(recover) - checking for panic & flush GA")
		gaMeasurement.Queue(measurement.NewTiming(time.Now().Sub(started)))
		if recovered := recover(); recovered != nil {
			messageText := fmt.Sprintf("Server error (panic): %v", recovered)
			logger.Criticalf(c, "Panic recovered: %s\n%s", messageText, debug.Stack())

			if gaMeasurement.QueueDepth() > 0 { // Zero if GA is disabled
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
			}

			if whc != nil {
				if whc.BotChatID() != nil {
					whc.Responder().SendMessage(c, whc.NewMessage(emoji.ERROR_ICON+" "+messageText), BotApiSendMessageOverResponse)
				}
			}
		} else if gaMeasurement.QueueDepth() > 0 { // Zero if GA is disabled
			logger.Debugf(c, "Flushing gaMeasurement (len(queue): %v)...", gaMeasurement.QueueDepth())
			if err = gaMeasurement.Flush(); err != nil {
				logger.Errorf(c, "Failed to send to GA: %v", err)
			} else {
				logger.Debugf(c, "Data sent to GA")
			}
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
			//logger.Debugf(c, "Closing BotChatStore...")
			chatEntity := whc.ChatEntity()
			if chatEntity != nil && chatEntity.GetPreferredLanguage() == "" {
				chatEntity.SetPreferredLanguage(whc.Locale().Code5)
			}
			if err := botCoreStores.BotChatStore.Close(); err != nil {
				logger.Errorf(c, "Failed to close BotChatStore: %v", err)
			}
		}
	}()

	for _, entryWithInputs := range entriesWithInputs {
		//logger.Infof(c, "Entry[%v]: %v, %v inputs", i, entryWithInputs.Entry.GetID(), len(entryWithInputs.Inputs))
		for j, input := range entryWithInputs.Inputs {
			inputType := input.InputType()
			switch inputType {
			case WebhookInputMessage, WebhookInputInlineQuery, WebhookInputCallbackQuery, WebhookInputChosenInlineResult:
				switch inputType {
				case WebhookInputMessage:
					sender := input.GetSender()
					logger.Infof(c, "User#%v(%v %v) text: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.InputMessage().Text())
				case WebhookInputCallbackQuery:
					callbackQuery := input.InputCallbackQuery()
					callbackData := callbackQuery.GetData()
					sender := input.GetSender()
					logger.Infof(c, "User#%v(%v %v) callback: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), callbackData)
				case WebhookInputInlineQuery:
					sender := input.GetSender()
					logger.Infof(c, "User#%v(%v %v) inline query: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.InputInlineQuery().GetQuery())
				case WebhookInputChosenInlineResult:
					sender := input.GetSender()
					logger.Infof(c, "User#%v(%v %v) choosen InlineMessageID: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.InputChosenInlineResult().GetInlineMessageID())
				}

				whc = webhookHandler.CreateWebhookContext(d.appContext, r, botContext, input, botCoreStores, gaMeasurement)
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
