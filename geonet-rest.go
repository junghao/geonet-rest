package main

import (
	"database/sql"
	"encoding/json"
	"github.com/GeoNet/geonet-rest/geojsonV1"
	"github.com/GeoNet/geonet-rest/jsonV1"
	"github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"time"
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
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "geonet-rest")
	if err == nil {
		log.SetOutput(logwriter)
	}

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

	// create an http client to share.
	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	// Create a router and subrouter so that all api requests are like http://server.com/api
	r := mux.NewRouter()
	// TODO - custom 404 handler.
	// r.NotFoundHandler = blah

	api := r.PathPrefix("/").Methods("GET").Subrouter()

	// All requests that have an Accept header that exactly matches geojsonV1.Accept will be sent to this router.
	v1 := api.Headers("Accept", geojsonV1.Accept).Subrouter()
	geojsonV1.Routes(v1, db, client)

	jv1 := api.Headers("Accept", jsonV1.Accept).Subrouter()
	jsonV1.RoutesHttp(jv1, client)

	// All requests that haven't exactly matched an earlier router Accept header are sent to this router.
	// It should route to the latest version of the API.
	geojsonV1.Routes(api, db, client)
	jsonV1.RoutesHttp(api, client)

	http.Handle("/", httpgzip.NewHandler(r))
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}
