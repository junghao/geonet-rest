package geojsonV1

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
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

// Valid is used to hold the response from GeoJSON validation.
type Valid struct {
	Status string
}

func validateGeoJSON(d []byte) (ok bool, err error) {
	client := &http.Client{}

	body := bytes.NewBuffer(d)

	r, err := client.Post("http://geojsonlint.com/validate", "application/vnd.geo+json", body)
	defer r.Body.Close()
	if err != nil {
		return false, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false, err
	}

	var v Valid

	err = json.Unmarshal(b, &v)
	if err != nil {
		return false, err
	}

	if v.Status != "ok" {
		return false, err
	}

	return true, err
}
