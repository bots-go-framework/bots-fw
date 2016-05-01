package kik

import (
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"strings"
	"fmt"
	"encoding/base64"
)

func KikWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func ConfigureKikHandler(w http.ResponseWriter, r *http.Request) {
	//This works
	// curl -H "Content-Type: application/json" -d '{"webhook": "https://debtstracker-io.appspot.com/bot/kik/webhook", "features": {"manuallySendReadReceipts": false, "receiveReadReceipts": false, "receiveDeliveryReceipts": false, "receiveIsTyping": false}}' -u 'debtstracker:1e296a7a-762a-4a00-9152-e9f410cacde1' 'https://api.kik.com/v1/config'

	//This does not
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	request, err := http.NewRequest("POST", "https://api.kik.com/v1/config", strings.NewReader(`{"webhook": "https://debtstracker-io.appspot.com/bot/kik/webhook", "features": {"manuallySendReadReceipts": false, "receiveReadReceipts": false, "receiveDeliveryReceipts": false, "receiveIsTyping": false}}`))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to create request: %v", err)))
	}
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", BOT_USERNAME, API_KEY)))))
	request.Header.Set("Content-Type", "application/json")

	res, err := client.Do(request)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to post settings to Kik: %v", err)))
	}
	body := make([]byte, res.ContentLength)
	_, err = res.Body.Read(body)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Failed to read response from Kik: %v", err)))
	}
	w.Write(body)
}