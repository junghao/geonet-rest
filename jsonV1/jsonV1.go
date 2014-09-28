// jsonV1 provides version=1 json services.
package jsonV1

import (
	"github.com/gorilla/mux"
	"net/http"
)

const Accept = "application/json; version 1;"

// makeHandler makes a handler that uses an http client.  Also sets http response headers.
func makeHandlerHttp(fn func(http.ResponseWriter, *http.Request, *http.Client), client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=60")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8; version 1;")
		fn(w, r, client)
	}
}

// Routes adds handlers to the mux.Router.
func RoutesHttp(r *mux.Router, client *http.Client) {
	// /news/geonet
	r.HandleFunc("/news/geonet", makeHandlerHttp(news, client))
}
