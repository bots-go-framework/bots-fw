package botsfw

import (
	"context"
	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/dal-go/dalgo/dal"
	"net/http"
)

type HttpRouter interface {
	Handle(method string, path string, handle http.HandlerFunc)
}

// WebhookHandler handles requests from a specific bot API
// This is implemented by different botsfw packages, e.g. https://github.com/bots-go-framework/bots-fw-telegram
// TODO: Simplify interface by decomposing it into smaller interfaces? Probably next method could/should be decoupled: CreateBotCoreStores()
type WebhookHandler interface {

	// RegisterHttpHandlers registers HTTP handlers for bot API
	RegisterHttpHandlers(driver WebhookDriver, botHost BotHost, router HttpRouter, pathPrefix string)

	// HandleWebhookRequest handles incoming webhook request
	HandleWebhookRequest(w http.ResponseWriter, r *http.Request)

	// GetBotContextAndInputs returns bot context and inputs for current request
	// It returns multiple inputs as some platforms (like Facebook Messenger)
	// may send multiple message in one request
	GetBotContextAndInputs(c context.Context, r *http.Request) (botContext *BotContext, entriesWithInputs []EntryInputs, err error)

	// CreateBotCoreStores TODO: should be deprecated after migration to dalgo
	//CreateBotCoreStores(appContext AppContext, r *http.Request) botsfwdal.DataAccess

	// CreateWebhookContext creates WebhookContext for current webhook request
	CreateWebhookContext(args CreateWebhookContextArgs) (WebhookContext, error)

	GetResponder(w http.ResponseWriter, whc WebhookContext) WebhookResponder
	HandleUnmatched(whc WebhookContext) (m MessageFromBot)
	//ProcessInput(input webhookInput, entry *WebhookEntry)
}

type CreateWebhookContextArgs struct {
	HttpRequest  *http.Request // TODO: Can we get rid of it? Needed for botHost.GetHTTPClient()
	AppContext   AppContext
	BotContext   BotContext
	WebhookInput botinput.WebhookInput
	Db           dal.DB
}

func NewCreateWebhookContextArgs(
	httpRequest *http.Request,
	appContext AppContext,
	botContext BotContext,
	webhookInput botinput.WebhookInput,
	db dal.DB,
) CreateWebhookContextArgs {
	return CreateWebhookContextArgs{
		HttpRequest:  httpRequest,
		AppContext:   appContext,
		BotContext:   botContext,
		WebhookInput: webhookInput,
		Db:           db,
	}
}
