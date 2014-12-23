package main

import (
	"database/sql"
	"github.com/GeoNet/app/cfg"
	"github.com/GeoNet/app/web"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

var (
	config = cfg.Load("geonet-rest") // this will be loaded before all init() func are called.
	db     *sql.DB                   // shared DB connection pool
	client *http.Client              // shared http client
)

var header = web.Header{
	Cache:     web.MaxAge10,
	Surrogate: web.MaxAge10,
	Vary:      "Accept",
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", config.Postgres())
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(config.DataBase.MaxIdleConns)
	db.SetMaxOpenConns(config.DataBase.MaxOpenConns)

	if err = db.Ping(); err != nil {
		log.Println("ERROR: problem pinging DB - is it up and contactable? 500s will be served")
	}

	// create an http client to share.
	timeout := time.Duration(5 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	return header.GetGzip(mux)
}