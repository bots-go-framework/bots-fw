package skype

import (
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"io"
	"net/http"
)

// WebhookHandler is handler of Skype API webhooks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, "FbmWebhookHandler")
	if r.Method == http.MethodGet {
		q := r.URL.Query()
		if q.Get("hub.verify_token") == "d6087a01-c728-4fdf-983c-1695d76236dc" {
			_, _ = w.Write([]byte(q.Get("hub.challenge")))
		} else {
			_, _ = w.Write([]byte("Error, wrong validation token"))
		}
	} else if r.Method == http.MethodPost {
		bytes, _ := io.ReadAll(r.Body)
		log.Infof(c, "request.BODY: %v", string(bytes))
		//w.WriteHeader(http.StatusNotImplemented)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
