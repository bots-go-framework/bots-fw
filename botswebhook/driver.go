package botswebhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsdal"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/gamp"
	"github.com/strongo/logus"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// ErrorIcon is used to report errors to user
var ErrorIcon = "🚨"

// BotDriver keeps information about bots and map requests to appropriate handlers
type BotDriver struct {
	Analytics       AnalyticsSettings
	botHost         botsfw.BotHost
	panicTextFooter string
}

var _ botsfw.WebhookDriver = (*BotDriver)(nil) // Ensure BotDriver is implementing interface WebhookDriver

// AnalyticsSettings keeps data for Google Analytics
type AnalyticsSettings struct {
	GaTrackingID string // TODO: Refactor to list of analytics providers
	Enabled      func(r *http.Request) bool
}

// NewBotDriver registers new bot driver (TODO: describe why we need it)
func NewBotDriver(gaSettings AnalyticsSettings, botHost botsfw.BotHost, panicTextFooter string) BotDriver {
	if botHost == nil {
		panic("required argument botHost == nil")
	}
	return BotDriver{
		Analytics:       gaSettings,
		botHost:         botHost,
		panicTextFooter: panicTextFooter,
	}
}

// RegisterWebhookHandlers adds handlers to a bot driver
func (d BotDriver) RegisterWebhookHandlers(httpRouter botsfw.HttpRouter, pathPrefix string, webhookHandlers ...botsfw.WebhookHandler) {
	for _, webhookHandler := range webhookHandlers {
		webhookHandler.RegisterHttpHandlers(d, d.botHost, httpRouter, pathPrefix)
	}
}

// HandleWebhook takes and HTTP request and process it
func (d BotDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler botsfw.WebhookHandler) {

	ctx := d.botHost.Context(r)

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

	// A bot can receiver multiple messages in a single request
	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(ctx, r)

	if d.invalidContextOrInputs(ctx, w, r, botContext, entriesWithInputs, err) {
		return
	}

	if len(entriesWithInputs) > 1 {
		log.Debugf(ctx, "BotDriver.HandleWebhook() => botCode=%v, len(entriesWithInputs): %d", botContext.BotSettings.Code, len(entriesWithInputs))
	}

	//botCoreStores := webhookHandler.CreateBotCoreStores(d.appContext, r)
	//defer func() {
	//	if whc != nil { // TODO: How do deal with Facebook multiple entries per request?
	//		//log.Debugf(c, "Closing BotChatStore...")
	//		//chatData := whc.ChatData()
	//		//if chatData != nil && chatData.GetPreferredLanguage() == "" {
	//		//	chatData.SetPreferredLanguage(whc.DefaultLocale().Code5)
	//		//}
	//	}
	//}()

	handleErrorAndReturnHttpError := func(err error, message string) {
		logus.Errorf(ctx, "%s: %v", message, err)
		errText := fmt.Sprintf("%s: %s: %v", http.StatusText(http.StatusInternalServerError), message, err)
		http.Error(w, errText, http.StatusInternalServerError)
	}

	handleErrorAndReturnHttpOK := func(err error, message string) {
		logus.Errorf(ctx, "%s: %v\nHTTP will return status OK", message, err)
		w.WriteHeader(http.StatusOK)
	}

	for _, entryWithInputs := range entriesWithInputs {
		for i, input := range entryWithInputs.Inputs {
			var handleError func(err error, message string)
			if input.InputType() == botinput.WebhookInputCallbackQuery {
				handleError = handleErrorAndReturnHttpOK
			} else {
				handleError = handleErrorAndReturnHttpError
			}
			if err = d.processWebhookInput(ctx, w, r, webhookHandler, botContext, i, input, handleError); err != nil {
				log.Errorf(ctx, "Failed to process input[%v]: %v", i, err)
			}
		}
	}
}

var isMissingGATrackingAlreadyReported bool

func (d BotDriver) processWebhookInput(
	ctx context.Context,
	w http.ResponseWriter, r *http.Request, webhookHandler botsfw.WebhookHandler,
	botContext *botsfw.BotContext,
	i int,
	input botinput.WebhookInput,
	handleError func(err error, message string),
) (
	err error,
) {
	var (
		whc               botsfw.WebhookContext // TODO: How do deal with Facebook multiple entries per request?
		measurementSender *gamp.BufferedClient
	)

	// Initiate Google Analytics Measurement API client
	sendStats := d.Analytics.Enabled != nil && d.Analytics.Enabled(r) ||
		botContext.BotSettings.Env == botsfw.EnvProduction

	if sendStats {
		if d.Analytics.GaTrackingID == "" {
			sendStats = false
			if !isMissingGATrackingAlreadyReported {
				log.Warningf(ctx, "driver.Analytics.GaTrackingID is not set")
			}
		} else {
			botHost := botContext.BotHost
			measurementSender = gamp.NewBufferedClient("", botHost.GetHTTPClient(ctx), func(err error) {
				log.Errorf(ctx, "Failed to log to GA: %v", err)
			})
		}
	} else {
		log.Debugf(ctx, "botContext.BotSettings.Env=%s, sendStats=%t", botContext.BotSettings.Env, sendStats)
	}

	started := time.Now()

	defer func() {
		log.Debugf(ctx, "driver.deferred(recover) - checking for panic & flush GA")
		if sendStats {
			timing := gamp.NewTiming(time.Since(started))
			timing.TrackingID = d.Analytics.GaTrackingID // TODO: What to do if different FB bots have different Tacking IDs? Can FB handler get messages for different bots? If not (what probably is the case) can we get ID from bot settings instead of driver?
			if err := measurementSender.Queue(timing); err != nil {
				log.Errorf(ctx, "Failed to log timing to GA: %v", err)
			}
		}

		reportError := func(recovered interface{}) {
			messageText := fmt.Sprintf("Server error (panic): %v\n\n%v", recovered, d.panicTextFooter)
			stack := string(debug.Stack())
			log.Criticalf(ctx, "Panic recovered: %s\n%s", messageText, stack)

			if sendStats { // Zero if GA is disabled
				d.reportErrorToGA(ctx, whc, measurementSender, messageText)
			}

			if whc != nil {
				var chatID string
				if chatID, err = whc.Input().BotChatID(); err == nil && chatID != "" {
					if responder := whc.Responder(); responder != nil {
						if _, err = responder.SendMessage(ctx, whc.NewMessage(ErrorIcon+" "+messageText), botsfw.BotAPISendMessageOverResponse); err != nil {
							log.Errorf(ctx, fmt.Errorf("failed to report error to user: %w", err).Error())
						}
					}
				}
			}
		}

		if recovered := recover(); recovered != nil {
			reportError(recovered)
		} else if sendStats {
			//log.Debugf(ctx, "Flushing GA...")
			if err = measurementSender.Flush(); err != nil {
				log.Warningf(ctx, "Failed to flush to GA: %v", err)
			} else if queueDepth := measurementSender.QueueDepth(); queueDepth > 0 {
				log.Debugf(ctx, "Sent to GA: %v items", queueDepth)
			}
		} else {
			log.Debugf(ctx, "GA: sendStats=false")
		}
	}()

	if input == nil {
		panic(fmt.Sprintf("entryWithInputs.Inputs[%d] == nil", i))
	}
	d.logInput(ctx, i, input)
	var db dal.DB
	if db, err = botContext.BotSettings.GetDatabase(ctx); err != nil {
		err = fmt.Errorf("failed to get bot database: %w", err)
		return
	}

	whcArgs := botsfw.NewCreateWebhookContextArgs(r, botContext.AppContext, *botContext, input, db, measurementSender)
	if whc, err = webhookHandler.CreateWebhookContext(whcArgs); err != nil {
		handleError(err, "Failed to create WebhookContext")
		return
	}
	chatData := whc.ChatData()

	if chatData != nil && chatData.GetAppUserID() == "" {
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {

			recordsToInsert := make([]dal.Record, 0)

			// chatData can be nil for inline requests
			// TODO: Should we try to deduct chat ID from user ID for inline queries inside a bot chat for "chat_type": "sender"?

			platformID := whc.BotPlatform().ID()
			botID := whc.GetBotCode()
			appContext := whc.AppContext()
			var appUser record.DataWithID[string, botsfwmodels.AppUserData]
			var botUser botsdal.BotUser
			bot := botsdal.Bot{
				Platform: botsfwconst.Platform(platformID),
				ID:       botID,
				User:     whc.Input().GetSender(),
			}
			if appUser, botUser, err = appContext.CreateAppUserFromBotUser(ctx, tx, bot); err != nil {
				return
			}
			if appUser.Record != nil {
				recordsToInsert = append(recordsToInsert, appUser.Record)
			}
			if botUser.Record != nil {
				recordsToInsert = append(recordsToInsert, botUser.Record)
			}

			chatData.SetAppUserID(appUser.ID)

			for _, recordToInsert := range recordsToInsert {
				if err = tx.Insert(ctx, recordToInsert); err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			handleError(err, fmt.Sprintf("Failed to run transaction for entriesWithInputs[%d]", i))
			return
		}
	}

	responder := webhookHandler.GetResponder(w, whc) // TODO: Move inside webhookHandler.CreateWebhookContext()?
	router := botContext.BotSettings.Profile.Router()

	if err = router.Dispatch(webhookHandler, responder, whc); err != nil {
		handleError(err, "Failed to dispatch")
		return
	}

	return
}

func (BotDriver) invalidContextOrInputs(c context.Context, w http.ResponseWriter, r *http.Request, botContext *botsfw.BotContext, entriesWithInputs []botsfw.EntryInputs, err error) bool {
	if err != nil {
		var errAuthFailed botsfw.ErrAuthFailed
		if errors.As(err, &errAuthFailed) {
			log.Warningf(c, "Auth failed: %v", err)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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
	case botsfw.EnvLocal:
		if !isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Local, host: %v", r.Host)
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
	case botsfw.EnvProduction:
		if isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Production, host: %v", r.Host)
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
	}

	return false
}

func isRunningLocally(host string) bool { // TODO(help-wanted): allow customization
	result := host == "localhost" ||
		strings.HasSuffix(host, ".ngrok.io") ||
		strings.HasSuffix(host, ".ngrok.dev") ||
		strings.HasSuffix(host, ".ngrok.app") ||
		strings.HasSuffix(host, ".ngrok-free.app")
	return result
}

func (BotDriver) reportErrorToGA(c context.Context, whc botsfw.WebhookContext, measurementSender *gamp.BufferedClient, messageText string) {
	log.Warningf(c, "reportErrorToGA() is temporary disabled")

	ga := whc.GA()
	if ga == nil {
		return
	}
	gaMessage := gamp.NewException(messageText, true)
	gaMessage.Common = ga.GaCommon()

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

func (BotDriver) logInput(c context.Context, i int, input botinput.WebhookInput) {
	sender := input.GetSender()
	prefix := fmt.Sprintf("BotUser#%v(%v %v)", sender.GetID(), sender.GetFirstName(), sender.GetLastName())
	switch input := input.(type) {
	case botinput.WebhookTextMessage:
		log.Debugf(c, "%s => text: %v", prefix, input.Text())
	case botinput.WebhookNewChatMembersMessage:
		newMembers := input.NewChatMembers()
		var b bytes.Buffer
		b.WriteString(fmt.Sprintf("NewChatMembers: %d", len(newMembers)))
		for i, member := range newMembers {
			b.WriteString(fmt.Sprintf("\t%d: (%v) - %v %v", i+1, member.GetUserName(), member.GetFirstName(), member.GetLastName()))
		}
		log.Debugf(c, b.String())
	case botinput.WebhookContactMessage:
		log.Debugf(c, "%s => Contact(botUserID=%s, firstName=%s)", prefix, input.GetBotUserID(), input.GetFirstName())
	case botinput.WebhookCallbackQuery:
		callbackData := input.GetData()
		log.Debugf(c, "%s => callback: %v", prefix, callbackData)
	case botinput.WebhookInlineQuery:
		log.Debugf(c, "%s => inline query: %v", prefix, input.GetQuery())
	case botinput.WebhookChosenInlineResult:
		log.Debugf(c, "%s => chosen InlineMessageID: %v", prefix, input.GetInlineMessageID())
	case botinput.WebhookReferralMessage:
		log.Debugf(c, "%s => text: %v", prefix, input.(botinput.WebhookTextMessage).Text())
	case botinput.WebhookSharedUsersMessage:
		sharedUsers := input.GetSharedUsers()
		log.Debugf(c, "%s => shared %d users", prefix, len(sharedUsers))
	default:
		log.Warningf(c, "unknown input[%v] type %T", i, input)
	}
}
