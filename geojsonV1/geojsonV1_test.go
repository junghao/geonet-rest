package geojsonV1

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
)

func serve(req *http.Request, res *httptest.ResponseRecorder) {
	db, err := sql.Open("postgres", "user=hazard_r password=test dbname=hazard sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	r := mux.NewRouter()
	Routes(r, db, client)
	r.ServeHTTP(res, req)
}
