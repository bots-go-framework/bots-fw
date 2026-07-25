package botswebhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bots-go-framework/bots-fw-store/botsfwstore"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/analytics"
	"github.com/strongo/logus"
)

// ErrorIcon is used to report errors to user
var ErrorIcon = "🚨"

const panicUserMessage = "Internal error. Please try again later."

// webhookDriver keeps information about bots and map requests to appropriate handlers
type webhookDriver struct {
	Analytics       AnalyticsSettings
	botHost         botsfw.BotHost
	panicTextFooter string
}

var _ botsfw.WebhookDriver = (*webhookDriver)(nil) // Ensure webhookDriver is implementing interface WebhookDriver

// AnalyticsSettings keeps data for Google Analytics
type AnalyticsSettings struct {
	GaTrackingID string // TODO: Refactor to list of analytics providers
	Enabled      func(r *http.Request) bool
}

// NewWebhookDriver registers new bot driver (TODO: describe why we need it)
func NewWebhookDriver(gaSettings AnalyticsSettings, botHost botsfw.BotHost, panicTextFooter string) botsfw.WebhookDriver {
	if botHost == nil {
		panic("required argument botHost == nil")
	}
	return webhookDriver{
		Analytics:       gaSettings,
		botHost:         botHost,
		panicTextFooter: panicTextFooter,
	}
}

// RegisterWebhookHandlers adds handlers to a bot driver
func (d webhookDriver) RegisterWebhookHandlers(httpRouter botsfw.HttpRouter, pathPrefix string, webhookHandlers ...botsfw.WebhookHandler) {
	for _, webhookHandler := range webhookHandlers {
		webhookHandler.RegisterHttpHandlers(d, d.botHost, httpRouter, pathPrefix)
	}
}

// HandleWebhook takes and HTTP request and process it
func (d webhookDriver) HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler botsfw.WebhookHandler) {
	//log.Debugf(c, "webhookDriver.HandleWebhook()")
	if w == nil {
		panic("Parameter 'w http.ResponseWriter' is nil")
	}
	if r == nil {
		panic("Parameter 'r *http.Request' is nil")
	}
	if webhookHandler == nil {
		panic("Parameter 'webhookHandler WebhookHandler' is nil")
	}

	response := newWebhookResponse(w)
	w = response.writer
	ctx := r.Context()
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Criticalf(ctx, "Panic recovered while handling webhook: %v\n\nStack trace:\n%s", recovered, debug.Stack())
			d.handleProcessingError(ctx, response, fmt.Errorf("panic while handling webhook: %v", recovered), "Panic while handling webhook")
		}
	}()

	if r.ContentLength > MaxWebhookRequestBodyBytes {
		log.Warningf(ctx, "webhook request body exceeds limit: content_length=%d, limit=%d", r.ContentLength, MaxWebhookRequestBodyBytes)
		response.writeError(http.StatusRequestEntityTooLarge)
		return
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, MaxWebhookRequestBodyBytes)
	}

	ctx = d.botHost.Context(r)

	// A bot can receiver multiple messages in a single request
	botContext, entriesWithInputs, err := webhookHandler.GetBotContextAndInputs(ctx, r)

	if d.invalidContextOrInputs(ctx, response, r, botContext, entriesWithInputs, err) {
		return
	}

	if len(entriesWithInputs) > 1 {
		log.Debugf(ctx, "webhookDriver.HandleWebhook() => botCode=%v, len(entriesWithInputs): %d", botContext.BotSettings.Code, len(entriesWithInputs))
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

	handleError := func(err error, message string) {
		d.handleProcessingError(ctx, response, err, message)
	}

	for _, entryWithInputs := range entriesWithInputs {
		if err = d.processWebhookEntry(ctx, response, r, webhookHandler, botContext, entryWithInputs, handleError); err != nil {
			log.Errorf(ctx, "Failed to process webhook entry: %v", err)
		}
	}
}

func (webhookDriver) handleProcessingError(ctx context.Context, response *webhookResponse, err error, operation string) {
	if response.isCommitted() {
		logus.Errorf(ctx, "%s: %v; response already committed", operation, err)
		return
	}
	logus.Errorf(ctx, "%s: %v", operation, err)
	response.writeError(http.StatusInternalServerError)
}

// processWebhookEntry leases one provider delivery and processes every input it
// contains before completing it. An entry is the provider's atomic retry unit;
// claiming per input would incorrectly suppress sibling inputs with the same
// update ID.
func (d webhookDriver) processWebhookEntry(
	ctx context.Context,
	response *webhookResponse, r *http.Request, webhookHandler botsfw.WebhookHandler,
	botContext *botsfw.BotContext,
	entryWithInputs botinput.EntryInputs,
	handleError func(error, string),
) (err error) {
	updateID, hasUpdateID := webhookUpdateID(entryWithInputs.Entry)
	store := botContext.BotSettings.Store
	if store == nil {
		err = fmt.Errorf("bot %q has no state store", botContext.BotSettings.Code)
		handleError(err, "Failed to prepare webhook state")
		return
	}

	var (
		updateKey botsfwstore.WebhookUpdateKey
		leaseID   string
	)
	if hasUpdateID {
		updateKey = botsfwstore.WebhookUpdateKey{PlatformID: string(botContext.BotSettings.Platform), BotID: botContext.BotSettings.Code, UpdateID: updateID}
		claim, claimErr := store.ClaimWebhookUpdate(ctx, updateKey, time.Now().UTC().Add(2*time.Minute))
		if claimErr != nil {
			err = fmt.Errorf("claim webhook update: %w", claimErr)
			handleError(err, "Failed to claim webhook update")
			return
		}
		switch claim.Status {
		case botsfwstore.WebhookUpdateClaimCompleted:
			if !response.isCommitted() {
				response.writer.WriteHeader(http.StatusOK)
			}
			return nil
		case botsfwstore.WebhookUpdateClaimLeased:
			response.writeError(http.StatusServiceUnavailable)
			return nil
		case botsfwstore.WebhookUpdateClaimAcquired:
			leaseID = claim.LeaseID
		default:
			err = fmt.Errorf("claim webhook update returned unknown status %q", claim.Status)
			handleError(err, "Failed to claim webhook update")
			return
		}
		defer func() {
			if err != nil && leaseID != "" {
				if failErr := store.FailWebhookUpdate(ctx, updateKey, leaseID, botsfwstore.WebhookUpdateFailureProcessing); failErr != nil {
					log.Errorf(ctx, "failed to mark webhook update as failed: %v", failErr)
				}
			}
		}()
	}

	for i, input := range entryWithInputs.Inputs {
		if err = d.processWebhookInput(ctx, response.writer, r, webhookHandler, botContext, i, input, handleError); err != nil {
			return fmt.Errorf("process input[%d]: %w", i, err)
		}
	}

	if leaseID != "" {
		if err = store.CompleteWebhookUpdate(ctx, updateKey, leaseID); err != nil {
			err = fmt.Errorf("complete webhook update: %w", err)
			handleError(err, "Failed to complete webhook update")
			return
		}
	}
	return nil
}

func webhookUpdateID(entry botinput.Entry) (string, bool) {
	if entry == nil {
		return "", false
	}
	if durableEntry, ok := entry.(botinput.DurableWebhookEntry); ok {
		id, ok := durableEntry.WebhookUpdateID()
		id = strings.TrimSpace(id)
		if !ok || id == "" {
			return "", false
		}
		return id, true
	}

	// Compatibility fallback for adapters that predate DurableWebhookEntry.
	// It deliberately rejects absent and zero IDs, both of which conventionally
	// mean the provider did not attach a durable delivery identifier.
	switch id := entry.GetID().(type) {
	case nil:
		return "", false
	case string:
		id = strings.TrimSpace(id)
		return id, id != ""
	case int:
		if id == 0 {
			return "", false
		}
	case int32:
		if id == 0 {
			return "", false
		}
	case int64:
		if id == 0 {
			return "", false
		}
	case uint:
		if id == 0 {
			return "", false
		}
	case uint32:
		if id == 0 {
			return "", false
		}
	case uint64:
		if id == 0 {
			return "", false
		}
	}
	return fmt.Sprint(entry.GetID()), true
}

func (d webhookDriver) processWebhookInput(
	ctx context.Context,
	w http.ResponseWriter, r *http.Request, webhookHandler botsfw.WebhookHandler,
	botContext *botsfw.BotContext,
	i int,
	input botinput.InputMessage,
	handleError func(err error, message string),
) (
	err error,
) {
	var (
		whc botsfw.WebhookContext // TODO: How do deal with Facebook multiple entries per request?
	)

	defer func() {
		log.Debugf(ctx, "driver.deferred(recover) - checking for panic & flush GA")

		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic while processing webhook input: %v", recovered)
			stack := string(debug.Stack())
			messageText := fmt.Sprintf("Panic: %v\n\nStack trace:\n%s\n\n%s", recovered, stack, d.panicTextFooter)
			log.Criticalf(ctx, "Panic recovered: %s", messageText)

			runPanicRecoveryStep(ctx, "write panic response", func() {
				handleError(err, "Panic while processing webhook input")
			})

			const maxLen = 3 * 1024
			if len(messageText) > maxLen {
				messageText = messageText[:maxLen] + fmt.Sprintf("\n\n...\n\nText truncated at %dKB", maxLen/1024)
			}

			// Initiate Google Analytics Measurement API client

			var analyticsEnabled bool
			runPanicRecoveryStep(ctx, "check analytics configuration", func() {
				analyticsEnabled = d.Analytics.Enabled != nil && d.Analytics.Enabled(r) || botContext.BotSettings.Env == botsfw.EnvProduction
			})
			if analyticsEnabled {
				if whc != nil {
					runPanicRecoveryStep(ctx, "report panic to analytics", func() {
						d.reportPanicToAnalytics(ctx, whc, messageText)
					})
				} else {
					log.Warningf(ctx, "not reporting panic to analytics: webhook context is unavailable")
				}
			} else {
				log.Debugf(ctx, "botContext.BotSettings.Env=%s, analyticsEnabled=%t", botContext.BotSettings.Env, analyticsEnabled)
			}

			if whc != nil {
				runPanicRecoveryStep(ctx, "report panic to user", func() {
					if chatID, chatIDErr := whc.Input().BotChatID(); chatIDErr == nil && chatID != "" {
						if responder := whc.Responder(); responder != nil {
							m := whc.NewMessage(ErrorIcon + " " + panicUserMessage)
							if _, sendErr := botsfw.SendMessageThroughGate(ctx, responder, m, botsfw.BotAPISendMessageOverHTTPS); sendErr != nil {
								if botsfw.IsSendNotPermitted(sendErr) {
									// A gated platform will not accept an unsolicited
									// message here. Reporting the panic to the user is
									// best-effort; the panic is already logged and sent
									// to analytics above.
									log.Warningf(ctx, "not reporting error to user: %v", sendErr)
								} else {
									log.Errorf(ctx, fmt.Errorf("failed to report error to user: %w", sendErr).Error())
								}
							}
						}
					}
				})
			}
		}
	}()

	if input == nil {
		panic(fmt.Sprintf("entryWithInputs.Inputs[%d] == nil", i))
	}
	d.logInput(ctx, i, input)
	store := botContext.BotSettings.Store
	if store == nil {
		err = fmt.Errorf("bot %q has no state store", botContext.BotSettings.Code)
		handleError(err, "Failed to prepare webhook state")
		return
	}
	whcArgs := botsfw.NewCreateWebhookContextArgs(r, botContext.AppContext, *botContext, input, store)
	if whc, err = webhookHandler.CreateWebhookContext(whcArgs); err != nil {
		handleError(err, "Failed to create WebhookContext")
		return
	}

	// Identity resolution and its persistence happen inside the injected store,
	// before dispatch. Router actions and outgoing messages remain outside that
	// persistence operation.
	_ = whc.ChatData()

	responder := webhookHandler.GetResponder(w, whc) // TODO: Move inside webhookHandler.CreateWebhookContext()?
	routerResponder := responder
	if provider, ok := webhookHandler.(botsfw.RouterResponderProvider); ok {
		routerResponder = provider.GetRouterResponder(whc, responder)
	}
	router := botContext.BotSettings.Profile.Router()

	if err = router.Dispatch(webhookHandler, routerResponder, whc); err != nil {
		handleError(err, "Failed to dispatch")
		return
	}
	return
}

func runPanicRecoveryStep(ctx context.Context, operation string, action func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Criticalf(ctx, "panic recovery step failed: operation=%q, panic_type=%T", operation, recovered)
		}
	}()
	action()
}

func (d webhookDriver) invalidContextOrInputs(c context.Context, response *webhookResponse, r *http.Request, botContext *botsfw.BotContext, entriesWithInputs []botinput.EntryInputs, err error) bool {
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			log.Warningf(c, "webhook request body exceeds limit: limit=%d", maxBytesError.Limit)
			response.writeError(http.StatusRequestEntityTooLarge)
			return true
		}
		var errAuthFailed botsfw.ErrAuthFailed
		if errors.As(err, &errAuthFailed) {
			log.Warningf(c, "Auth failed: %v", err)
			response.writeError(http.StatusForbidden)
			return true
		}
		d.handleProcessingError(c, response, err, "Failed to parse webhook request")
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
		response.writeError(http.StatusInternalServerError)
		return true
	} else if entriesWithInputs == nil {
		log.Errorf(c, "entriesWithInputs == nil")
		response.writeError(http.StatusInternalServerError)
		return true
	}

	switch botContext.BotSettings.Env {
	case botsfw.EnvLocal:
		if !isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Local, host: %v", r.Host)
			response.writeError(http.StatusBadRequest)
			return true
		}
	case botsfw.EnvProduction:
		if isRunningLocally(r.Host) {
			log.Warningf(c, "whc.GetBotSettings().Mode == Production, host: %v", r.Host)
			response.writeError(http.StatusBadRequest)
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

func (webhookDriver) reportPanicToAnalytics(c context.Context, whc botsfw.WebhookContext, messageText string) {
	log.Warningf(c, "reportPanicToAnalytics() is temporary disabled")
	err := fmt.Errorf("%s", messageText)
	msg := analytics.NewErrorMessage(err) // TODO: replace with analytics.NewPanicMessage()
	whc.Analytics().Enqueue(msg)
}

func (webhookDriver) logInput(c context.Context, i int, input botinput.InputMessage) {
	sender := input.GetSender()
	prefix := fmt.Sprintf("BotUser#%v(%v %v)", sender.GetID(), sender.GetFirstName(), sender.GetLastName())
	switch input := input.(type) {
	case botinput.TextMessage:
		log.Debugf(c, "%s => text: %v", prefix, input.Text())
	case botinput.LocationMessage:
		log.Debugf(c, "%s => location: %d:%d", prefix, input.GetLatitude(), input.GetLongitude())
	case botinput.NewChatMembersMessage:
		newMembers := input.NewChatMembers()
		var b bytes.Buffer
		fmt.Fprintf(&b, "NewChatMembers: %d", len(newMembers))
		for i, member := range newMembers {
			fmt.Fprintf(&b, "\t%d: (%v) - %v %v", i+1, member.GetUserName(), member.GetFirstName(), member.GetLastName())
		}
		log.Debugf(c, b.String())
	case botinput.ContactMessage:
		log.Debugf(c, "%s => Contact(botUserID=%s, firstName=%s)", prefix, input.GetBotUserID(), input.GetFirstName())
	case botinput.CallbackQuery:
		callbackData := input.GetData()
		log.Debugf(c, "%s => callback: %v", prefix, callbackData)
	case botinput.InlineQuery:
		log.Debugf(c, "%s => inline query: %v", prefix, input.GetQuery())
	case botinput.ChosenInlineResult:
		log.Debugf(c, "%s => chosen InlineMessageID: %v", prefix, input.GetInlineMessageID())
	case botinput.ReferralMessage:
		log.Debugf(c, "%s => referral: type=%v source=%v refData=%v", prefix, input.Type(), input.Source(), input.RefData())
	case botinput.SharedUsersMessage:
		sharedUsers := input.GetSharedUsers()
		log.Debugf(c, "%s => shared %d users", prefix, len(sharedUsers))
	default:
		log.Warningf(c, "unknown input[%v] type %T", i, input)
	}
}
