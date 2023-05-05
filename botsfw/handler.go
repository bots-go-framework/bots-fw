package botsfw

import (
	"context"
	"net/http"
)

type HttpRouter interface {
	Handle(method string, path string, handle http.HandlerFunc)
}

// WebhookHandler handles requests from a specific bot API
// TODO: Simplify interface by decomposing it into smaller interfaces? Probably next method could/should be decoupled: CreateBotCoreStores()
type WebhookHandler interface {
	RegisterHttpHandlers(driver WebhookDriver, botHost BotHost, router HttpRouter, pathPrefix string)
	HandleWebhookRequest(w http.ResponseWriter, r *http.Request)
	GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *BotContext, entriesWithInputs []EntryInputs, err error)
	CreateBotCoreStores(appContext BotAppContext, r *http.Request) BotCoreStores
	CreateWebhookContext(appContext BotAppContext, r *http.Request, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores, gaMeasurement GaQueuer) WebhookContext //TODO: Can we get rid of http.Request? Needed for botHost.GetHTTPClient()
	GetResponder(w http.ResponseWriter, whc WebhookContext) WebhookResponder
	HandleUnmatched(whc WebhookContext) (m MessageFromBot)
	//ProcessInput(input webhookInput, entry *WebhookEntry)
}
