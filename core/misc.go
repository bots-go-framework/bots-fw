package bots

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("Pong"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
