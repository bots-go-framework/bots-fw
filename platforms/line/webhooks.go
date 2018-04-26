package line

import (
	"net/http"
)

// WebhookHandler is handler of Line API webhooks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
