package jsonV1

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
)

func serve(req *http.Request, res *httptest.ResponseRecorder) {
	client := &http.Client{}

	r := mux.NewRouter()
	RoutesHttp(r, client)
	r.ServeHTTP(res, req)
}
