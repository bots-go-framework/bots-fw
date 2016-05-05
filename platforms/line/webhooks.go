package line

import (
	"net/http"
)

func LineWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
