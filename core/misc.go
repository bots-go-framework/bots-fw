package bots

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("Pong"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
