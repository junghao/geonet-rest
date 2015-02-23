package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/http/httptest"
)

var ts *httptest.Server

// setup starts a db connection and test server then inits an http client.
func setup() {
	var err error
	db, err = sql.Open("postgres", config.DataBase.Postgres())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	ts = httptest.NewServer(handler())

	client = &http.Client{}
}

// teardown closes the db connection and  test server.  Defer this after setup() e.g.,
// ...
// setup()
// defer teardown()
func teardown() {
	ts.Close()
	db.Close()
}

// Valid is used to hold the response from GeoJSON validation.
type Valid struct {
	Status string
}
