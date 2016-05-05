package skype

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"io/ioutil"
	"net/http"
)

func SkypeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Infof(c, "FbmWebhookHandler")
	if r.Method == http.MethodGet {
		q := r.URL.Query()
		if q.Get("hub.verify_token") == "d6087a01-c728-4fdf-983c-1695d76236dc" {
			w.Write([]byte(q.Get("hub.challenge")))
		} else {
			w.Write([]byte("Error, wrong validation token"))
		}
	} else if r.Method == http.MethodPost {
		bytes, _ := ioutil.ReadAll(r.Body)
		log.Infof(c, "request.BODY: %v", string(bytes))
		//w.WriteHeader(http.StatusNotImplemented)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
