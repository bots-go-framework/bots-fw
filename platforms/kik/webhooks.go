package kik

import (
	"encoding/base64"
	"fmt"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"strings"
)

// WebhookHandler handles Kik requests
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Errorf(r.Context(), "Failed to write to response: %v", err)
	}
}

// ConfigureKikHandler configures kik bot
func ConfigureKikHandler(w http.ResponseWriter, r *http.Request) {
	//This works
	// curl -H "Content-Type: application/json" -d '{"webhook": "https://debtstracker-io.appspot.com/bot/kik/webhook", "features": {"manuallySendReadReceipts": false, "receiveReadReceipts": false, "receiveDeliveryReceipts": false, "receiveIsTyping": false}}' -u 'debtstracker:1e296a7a-762a-4a00-9152-e9f410cacde1' 'https://api.kik.com/v1/config'

	//This does not
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	request, err := http.NewRequest("POST", "https://api.kik.com/v1/config", strings.NewReader(`{"webhook": "https://debtstracker-io.appspot.com/bot/kik/webhook", "features": {"manuallySendReadReceipts": false, "receiveReadReceipts": false, "receiveDeliveryReceipts": false, "receiveIsTyping": false}}`))
	if err != nil {
		if _, err2 := w.Write([]byte(fmt.Sprintf("Failed to create request: %v", err))); err2 != nil {
			log.Errorf(c, "Failed to write to response: %v", err2)
		}
	}
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", "BOT_USERNAME", "API_KEY")))))
	request.Header.Set("Content-Type", "application/json")

	res, err := client.Do(request)
	if err != nil {
		if _, err2 := w.Write([]byte(fmt.Sprintf("Failed to post settings to Kik: %v", err))); err2 != nil {
			log.Errorf(c, "Failed to write to response: %v", err2)
		}
	}
	body := make([]byte, res.ContentLength)

	if _, err = res.Body.Read(body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err2 := w.Write([]byte(fmt.Sprintf("Failed to read response from Kik: %v", err))); err2 != nil {
			log.Errorf(c, "Failed to write to response: %v", err2)
		}
		return
	}
	if _, err := w.Write(body); err != nil {
		log.Errorf(c, "Failed to write to response: %v", err)
	}
}
