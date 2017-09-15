package bots

import (
	"net/http"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"github.com/julienschmidt/httprouter"
)

type WebhookHandler interface {
	RegisterWebhookHandler(driver WebhookDriver, botHost BotHost, router *httprouter.Router, pathPrefix string)
	HandleWebhookRequest(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *BotContext, entriesWithInputs []EntryInputs, err error)
	CreateBotCoreStores(appContext BotAppContext, r *http.Request) BotCoreStores
	CreateWebhookContext(appContext BotAppContext, r *http.Request, botContext BotContext, webhookInput WebhookInput, botCoreStores BotCoreStores, gaMeasurement *measurement.BufferedSender) WebhookContext //TODO: Can we get rid of http.Request? Needed for botHost.GetHttpClient()
	GetResponder(w http.ResponseWriter, whc WebhookContext) WebhookResponder
	//ProcessInput(input WebhookInput, entry *WebhookEntry)
}
