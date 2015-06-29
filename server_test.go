package main

import (
	"database/sql"
	"github.com/GeoNet/msg"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/http/httptest"
)

var ts *httptest.Server

// setup starts a db connection and test server then inits an http client.
func setup() {
	// load some test data.  Needs a write user.
	var err error
	config.DataBase.User = "hazard_w"
	db, err = sql.Open("postgres", config.DataBase.Postgres())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	// TODO add some more data and check the size of some region queries?
	q := msg.ReadSC3ML07("etc/test/files/2013p407387.xml")
	if err != nil {
		log.Fatal(err)
	}

	err = saveQuake(q)
	if err != nil {
		log.Fatal(err)
	}

	db.Close()

	// switch back to the correct user for the tests.
	// hazard_r can read haz and impact.
	config.DataBase.User = "hazard_r"
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
