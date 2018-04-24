package bots

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// PingHandler returns 'Pong' back to user
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("Pong"))
}

// NotFoundHandler returns HTTP status code 404
func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
