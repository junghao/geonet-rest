package main

import (
	"database/sql"
	"encoding/json"
	"github.com/GeoNet/geonet-rest/geojsonV1"
	"github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
)

var config Config

// Config is used to hold config parsed from geonet-rest.json
type Config struct {
	DataBase DataBase
	Server   Server
}

type DataBase struct {
	User, Password             string
	MaxOpenConns, MaxIdleConns int
}

type Server struct {
	Port string
}

// init loads configuration for this application.  It tries /etc/sysconfig/geonet-rest.json first and
// if that is not found it tries ./geonet-rest.json
func init() {
	f, err := ioutil.ReadFile("/etc/sysconfig/geonet-rest.json")
	if err != nil {
		f, err = ioutil.ReadFile("./geonet-rest.json")
		if err != nil {
			log.Fatal(err)
		}
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Fatal(err)
	}
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	db, err := sql.Open("postgres", "user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=hazard sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(config.DataBase.MaxIdleConns)
	db.SetMaxOpenConns(config.DataBase.MaxOpenConns)

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create a router and subrouter so that all api requests are like http://server.com/api
	r := mux.NewRouter()

	api := r.PathPrefix("/api/").Methods("GET").Subrouter()

	// All requests that have an Accept header that exactly matches geojsonV1.Accept will be sent to this router.
	v1 := api.Headers("Accept", geojsonV1.Accept).Subrouter()
	geojsonV1.Routes(v1, db)

	// All requests that haven't exactly matched an earlier router Accept header are sent to this router.
	// It should route to the latest version of the API.
	geojsonV1.Routes(api, db)

	http.Handle("/", httpgzip.NewHandler(r))
	http.ListenAndServe(":"+config.Server.Port, nil)
}
