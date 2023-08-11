package botsfw

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/strongo/app"
	"github.com/strongo/gamp"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// ErrorIcon is used to report errors to user
var ErrorIcon = "🚨"

// WebhookDriver is doing initial request & final response processing.
// That includes logging, creating input messages in a general format, sending response.
type WebhookDriver interface {
	RegisterWebhookHandlers(httpRouter HttpRouter, pathPrefix string, webhookHandlers ...WebhookHandler)
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}

// BotDriver keeps information about bots and map requests to appropriate handlers
type BotDriver struct {
	Analytics  AnalyticsSettings
	botHost    BotHost
	appContext BotAppContext
	//router          *WebhooksRouter
	panicTextFooter string
}

var _ WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

// AnalyticsSettings keeps data for Google Analytics
type AnalyticsSettings struct {
	GaTrackingID string // TODO: Refactor to list of analytics providers
	Enabled      func(r *http.Request) bool
}

// NewBotDriver registers new bot driver (TODO: describe why we need it)
func NewBotDriver(gaSettings AnalyticsSettings, appContext BotAppContext, host BotHost, panicTextFooter string) BotDriver {
	if appContext.AppUserCollectionName() == "" {
		panic("appContext.AppUserCollectionName() is empty")
	}
	if host == nil {
		panic("BotHost == nil")
	}
	return BotDriver{
		Analytics:  gaSettings,
		appContext: appContext,
		botHost:    host,
		//router: router,
		panicTextFooter: panicTextFooter,
	}
}

// RegisterWebhookHandlers adds handlers to a bot driver
func (d BotDriver) RegisterWebhookHandlers(httpRouter HttpRouter, pathPrefix string, webhookHandlers ...WebhookHandler) {
	for _, webhookHandler := range webhookHandlers {
		webhookHandler.RegisterHttpHandlers(d, d.botHost, httpRouter, pathPrefix)
	}
}

// HandleWebhook takes and HTTP request and process it
func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler) {
	started := time.Now()
	c := d.botHost.Context(r)
	//log.Debugf(c, "BotDriver.HandleWebhook()")
	if w == nil {
		panic("Parameter 'w http.ResponseWriter' is nil")
	}
	if r == nil {
		panic("Parameter 'r *http.Request' is nil")
	}
	if webhookHandler == nil {
		panic("Parameter 'webhookHandler WebhookHandler' is nil")
	}

	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(c, r)

	if d.invalidContextOrInputs(c, w, r, botContext, entriesWithInputs, err) {
		return
	}

	log.Debugf(c, "BotDriver.HandleWebhook() => botCode=%v, len(entriesWithInputs): %d", botContext.BotSettings.Code, len(entriesWithInputs))

	var (
		whc               WebhookContext // TODO: How do deal with Facebook multiple entries per request?
		measurementSender *gamp.BufferedClient
	)

	var sendStats bool
	{ // Initiate Google Analytics Measurement API client

		if d.Analytics.Enabled == nil {
			sendStats = botContext.BotSettings.Env == strongo.EnvProduction
			//} else {
			//if sendStats = d.Analytics.Enabled(r); !sendStats {
			//
			//}
			//log.Debugf(c, "d.AnalyticsSettings.Enabled != nil, sendStats: %v", sendStats)
		}
		if sendStats {
			botHost := botContext.BotHost
			measurementSender = gamp.NewBufferedClient("", botHost.GetHTTPClient(c), func(err error) {
				log.Errorf(c, "Failed to log to GA: %v", err)
			})
		} else {
			envName, ok := strongo.EnvironmentNames[botContext.BotSettings.Env]
			if !ok {
				envName = "UNKNOWN"
			}
			log.Debugf(c, "d.Analytics.Enabled=%v, botContext.BotSettings.Env=%v:%v, sendStats=%v",
				d.Analytics.Enabled, botContext.BotSettings.Env, envName, sendStats)
		}
	}

	defer func() {
		log.Debugf(c, "driver.deferred(recover) - checking for panic & flush GA")
		if sendStats {
			if d.Analytics.GaTrackingID == "" {
				log.Warningf(c, "driver.Analytics.GaTrackingID is not set")
			} else {
				timing := gamp.NewTiming(time.Since(started))
				timing.TrackingID = d.Analytics.GaTrackingID // TODO: What to do if different FB bots have different Tacking IDs? Can FB handler get messages for different bots? If not (what probably is the case) can we get ID from bot settings instead of driver?
				if err := measurementSender.Queue(timing); err != nil {
					log.Errorf(c, "Failed to log timing to GA: %v", err)
				}
			}
		}

		reportError := func(recovered interface{}) {
			messageText := fmt.Sprintf("Server error (panic): %v\n\n%v", recovered, d.panicTextFooter)
			log.Criticalf(c, "Panic recovered: %s\n%s", messageText, debug.Stack())

			if sendStats { // Zero if GA is disabled
				d.reportErrorToGA(c, whc, measurementSender, messageText)
			}

			if whc != nil {
				if chatID, err := whc.BotChatID(); err == nil && chatID != "" {
					if responder := whc.Responder(); responder != nil {
						if _, err := responder.SendMessage(c, whc.NewMessage(ErrorIcon+" "+messageText), BotAPISendMessageOverResponse); err != nil {
							log.Errorf(c, fmt.Errorf("failed to report error to user: %w", err).Error())
						}
					}
				}
			}
		}

		if recovered := recover(); recovered != nil {
			reportError(recovered)
		} else if sendStats {
			log.Debugf(c, "Flushing GA...")
			if err = measurementSender.Flush(); err != nil {
				log.Warningf(c, "Failed to flush to GA: %v", err)
			} else {
				log.Debugf(c, "Sent to GA: %v items", measurementSender.QueueDepth())
			}
		} else {
			log.Debugf(c, "GA: sendStats=false")
		}
	}()

	if err != nil {
		log.Errorf(c, "Failed to create new WebhookContext: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	botCoreStores := webhookHandler.CreateBotCoreStores(d.appContext, r)
	defer func() {
		if whc != nil { // TODO: How do deal with Facebook multiple entries per request?
			//log.Debugf(c, "Closing BotChatStore...")
			//chatData := whc.ChatData()
			//if chatData != nil && chatData.GetPreferredLanguage() == "" {
			//	chatData.SetPreferredLanguage(whc.Locale().Code5)
			//}
			if botCoreStores != nil {
				if err := botCoreStores.Close(c); err != nil {
					log.Errorf(c, "Failed to close BotChatStore: %v", err)
					var m MessageFromBot
					m.Text = ErrorIcon + " ERROR: Service is temporary unavailable. Probably a global outage, status at https://status.cloud.google.com/"
					if _, err := whc.Responder().SendMessage(c, m, BotAPISendMessageOverHTTPS); err != nil {
						log.Errorf(c, "Failed to report outage: %v", err)
					}
				}
			}
		}
	}()

	for _, entryWithInputs := range entriesWithInputs {
		for i, input := range entryWithInputs.Inputs {
			if input == nil {
				panic(fmt.Sprintf("entryWithInputs.Inputs[%d] == nil", i))
			}
			d.logInput(c, i, input)
			whc = webhookHandler.CreateWebhookContext(d.appContext, r, *botContext, input, botCoreStores, measurementSender)
			responder := webhookHandler.GetResponder(w, whc) // TODO: Move inside webhookHandler.CreateWebhookContext()?
			botContext.BotSettings.Router.Dispatch(webhookHandler, responder, whc)
		}
	}
}

func (BotDriver) invalidContextOrInputs(c context.Context, w http.ResponseWriter, r *http.Request, botContext *BotContext, entriesWithInputs []EntryInputs, err error) bool {
	if err != nil {
		if _, ok := err.(ErrAuthFailed); ok {
			log.Warningf(c, "Auth failed: %v", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else if errors.Is(err, ErrNotImplemented) {
			log.Debugf(c, err.Error())
			w.WriteHeader(http.StatusNoContent)
			//http.Error(w, "", http.StatusOK) // TODO: Decide how to handle it properly, return http.StatusNotImplemented?
		} else if _, ok := err.(*json.SyntaxError); ok {
			log.Debugf(c, fmt.Errorf("request body is not valid JSON: %w", err).Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			log.Errorf(c, "Failed to call webhookHandler.GetBotContextAndInputs(router): %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return true
	}
	if botContext == nil {
		if entriesWithInputs == nil {
			log.Warningf(c, "botContext == nil, entriesWithInputs == nil")
		} else if len(entriesWithInputs) == 0 {
			log.Warningf(c, "botContext == nil, len(entriesWithInputs) == 0")
		} else {
			log.Errorf(c, "botContext == nil, len(entriesWithInputs) == %v", len(entriesWithInputs))
		}
		return true
	} else if entriesWithInputs == nil {
		log.Errorf(c, "entriesWithInputs == nil")
		return true
	}

	switch botContext.BotSettings.Env {
	case strongo.EnvLocal:
		if !isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Local, host: %v", r.Host)
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
	case strongo.EnvProduction:
		if isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Production, host: %v", r.Host)
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
	}

	return false
}

func isRunningLocally(host string) bool {
	result := host == "localhost" ||
		strings.HasSuffix(host, ".ngrok.io") ||
		strings.HasSuffix(host, ".ngrok.dev") ||
		strings.HasSuffix(host, ".ngrok.app") ||
		strings.HasSuffix(host, ".ngrok-free.app")
	return result
}

func (BotDriver) reportErrorToGA(c context.Context, whc WebhookContext, measurementSender *gamp.BufferedClient, messageText string) {
	log.Warningf(c, "reportErrorToGA() is temporary disabled")
	if c != nil { // TODO: Remove once fixed
		return
	}
	ga := whc.GA()
	gaMessage := gamp.NewException(messageText, true)

	if whc != nil { // TODO: How do deal with Facebook multiple entries per request?
		gaMessage.Common = ga.GaCommon()
	} else {
		gaMessage.Common.ClientID = "c7ea15eb-3333-4d47-a002-9d1a14996371" // TODO: move hardcoded value
		gaMessage.Common.DataSource = "bot-" + whc.BotPlatform().ID()
	}

	if err := ga.Queue(gaMessage); err != nil {
		log.Errorf(c, "Failed to queue exception message for GA: %v", err)
	} else {
		log.Debugf(c, "Exception message queued for GA.")
	}

	if err := measurementSender.Flush(); err != nil {
		log.Errorf(c, "Failed to flush GA buffer after exception: %v", err)
	} else {
		log.Debugf(c, "GA buffer flushed after exception")
	}
}

func (BotDriver) logInput(c context.Context, i int, input WebhookInput) {
	sender := input.GetSender()
	switch input := input.(type) {
	case WebhookTextMessage:
		log.Debugf(c, "BotUser#%v(%v %v) => text: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.Text())
	case WebhookNewChatMembersMessage:
		newMembers := input.NewChatMembers()
		var b bytes.Buffer
		b.WriteString(fmt.Sprintf("NewChatMembers: %d", len(newMembers)))
		for i, member := range newMembers {
			b.WriteString(fmt.Sprintf("\t%d: (%v) - %v %v", i+1, member.GetUserName(), member.GetFirstName(), member.GetLastName()))
		}
		log.Debugf(c, b.String())
	case WebhookContactMessage:
		log.Debugf(c, "BotUser#%v(%v %v) => Contact(name: %v|%v, phone number: %v)", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.FirstName(), input.LastName(), input.PhoneNumber())
	case WebhookCallbackQuery:
		callbackData := input.GetData()
		log.Debugf(c, "BotUser#%v(%v %v) => callback: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), callbackData)
	case WebhookInlineQuery:
		log.Debugf(c, "BotUser#%v(%v %v) => inline query: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.GetQuery())
	case WebhookChosenInlineResult:
		log.Debugf(c, "BotUser#%v(%v %v) => chosen InlineMessageID: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.GetInlineMessageID())
	case WebhookReferralMessage:
		log.Debugf(c, "BotUser#%v(%v %v) => text: %v", sender.GetID(), sender.GetFirstName(), sender.GetLastName(), input.(WebhookTextMessage).Text())
	default:
		log.Warningf(c, "Unhandled input[%v] type: %T", i, input)
	}
}
