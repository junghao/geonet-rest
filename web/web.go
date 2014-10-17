// web writes responses to clients.  Logs request and responses and use expvar to expose counters at /debug/vars
package web

import (
	"expvar"
	"log"
	"net/http"
)

var req = expvar.NewInt("requests")
var res = expvar.NewMap("responses")

func init() {
	res.Init()
	res.Add("2xx", 0)
	res.Add("4xx", 0)
	res.Add("5xx", 0)
}

// Win (200) - writes the content in b to the client.
func Win(w http.ResponseWriter, r *http.Request, b []byte) {
	// Haven't bothered logging sucesses.
	res.Add("2xx", 1)
	req.Add(1)
	w.Write(b)
}

// Nope (404) - whatever the client was looking for we haven't got it.  The message should try
// to explain why we couldn't find that thing that they was looking for.
func Nope(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 404")
	res.Add("4xx", 1)
	req.Add(1)
	http.Error(w, message, 404)
}

// Fail (500) - some sort of internal server error.
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 500")
	res.Add("5xx", 1)
	req.Add(1)
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", 500)
}
