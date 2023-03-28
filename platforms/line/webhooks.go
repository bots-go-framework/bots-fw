package line

import (
	"github.com/strongo/log"
	"net/http"
)

// WebhookHandler is handler of Line API webhooks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Errorf(r.Context(), "Failed to write to response: %v", err)
	}
}
