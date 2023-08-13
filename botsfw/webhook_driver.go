package botsfw

import "net/http"

// WebhookDriver is doing initial request & final response processing.
// That includes logging, creating input messages in a general format, sending response.
type WebhookDriver interface {
	RegisterWebhookHandlers(httpRouter HttpRouter, pathPrefix string, webhookHandlers ...WebhookHandler)
	HandleWebhook(w http.ResponseWriter, r *http.Request, webhookHandler WebhookHandler)
}
