package main

import (
	"database/sql"
	"encoding/json"
	"expvar"
	"github.com/daaku/go.httpgzip"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"time"
)

const (
	v1GeoJSON   = "application/vnd.geo+json;version=1"
	v1JSON      = "application/json;version=1"
	mlink       = "http://info.geonet.org.nz/m/view-rendered-page.action?abstractPageId="
	newsURL     = "http://info.geonet.org.nz/createrssfeed.action?types=blogpost&spaces=conf_all&title=GeoNet+News+RSS+Feed&labelString%3D&excludedSpaceKeys%3D&sort=created&maxResults=10&timeSpan=500&showContent=true&publicFeed=true&confirm=Create+RSS+Feed"
	feltURL     = "http://felt.geonet.org.nz/services/reports/"
	cacheShort  = "max-age=10"
	cacheMedium = "max-age=300"
	cacheLong   = "max-age=86400"
)

var (
	config  = initConfig()              // this will be loaded before all init() func are called.
	db      *sql.DB                     // shared DB connection pool
	client  *http.Client                // shared http client
	req     = expvar.NewInt("requests") // counters for expvar
	res     = expvar.NewMap("responses")
	resTime timer
)

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

// initConfig loads configuration for this application.  It tries /etc/sysconfig/geonet-rest.json first and
// if that is not found it tries ./geonet-rest.json
// If the config is succesfully loaded from /etc/sysconfig/geonet-rest.json then the logging
// switches to syslogging.
func initConfig() Config {
	f, err := ioutil.ReadFile("/etc/sysconfig/geonet-rest.json")
	if err != nil {
		log.Println("Could not load /etc/sysconfig/geonet-rest.json falling back to local file.")
		f, err = ioutil.ReadFile("./geonet-rest.json")
		if err != nil {
			log.Println("Problem loading ./geonet-rest.json - can't find any config.")
			log.Fatal(err)
		}
	} else {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "geonet-rest")
		if err != nil {
			log.Println("** logging to syslog **")
			log.SetOutput(logwriter)
		}
	}

	var d Config
	err = json.Unmarshal(f, &d)
	if err != nil {
		log.Println("Problem parsing config file.")
		log.Fatal(err)
	}

	return d
}

func init() {
	res.Init()
	res.Add("2xx", 0)
	res.Add("4xx", 0)
	res.Add("5xx", 0)

	resTime = timer{count: 0, time: 0, interval: 30 * time.Second, v: expvar.NewFloat("averageResonseTime")}
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", "connect_timeout=1 user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=hazard sslmode=disable")
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(config.DataBase.MaxIdleConns)
	db.SetMaxOpenConns(config.DataBase.MaxOpenConns)

	err = db.Ping()

	if err != nil {
		log.Println("Problem pinging DB - is it up and contactable.")
		log.Fatal(err)
	}

	// create an http client to share.
	timeout := time.Duration(5 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}

	go resTime.avg()

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", noRoute)
	mux.HandleFunc("/quake/", quakeRoutes)
	mux.HandleFunc("/quake", quakesRoutes)
	mux.HandleFunc("/region/", regionRoutes)
	mux.HandleFunc("/region", regionsRoutes)
	mux.HandleFunc("/felt/report", reportRoutes)
	mux.HandleFunc("/news/", newsRoutes)
	return get(httpgzip.NewHandler(mux))
}

func noRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		badRequest(w, r, "service not found.")
	case v1JSON:
		badRequest(w, r, "service not found.")
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.  Try using: "+v1GeoJSON)
	}
}

// get creates an http handler that only responds to http GET requests.  All other methods are an error.
// Sets a default Cache-Control and Surrogate-Control header.
// Increments the request counter.
func get(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req.Add(1)
		if r.Method == "GET" {
			defer resTime.track(time.Now(), "GET "+r.URL.RequestURI())
			w.Header().Set("Cache-Control", cacheShort)
			w.Header().Set("Surrogate-Control", cacheShort)
			w.Header().Add("Vary", "Accept")
			h.ServeHTTP(w, r)
			return
		}
		res.Add("4xx", 1)
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	})
}

// quakeRoutes handles requests  for single quakes e.g., /quake/2013p12345
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func quakeRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		quakeV1(w, r)
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.  Try using: "+v1GeoJSON)
	}
}

// quakesRoutes handles request that filter lists of quakes e.g., /quake?regionID=...
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func quakesRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		switch r.URL.Query().Get("regionIntensity") {
		case "":
			quakesV1(w, r)
		default:
			quakesRegionV1(w, r)
		}
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.  Try using: "+v1GeoJSON)
	}
}

// regionRoutes handles requests  for single regions e.g., /region/newzealand
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func regionRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		regionV1(w, r)
	default:
		notAcceptable(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// regionsRoutes handles request that filter lists of region e.g., /region?type=quake.
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func regionsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		regionsV1(w, r)
	default:
		notAcceptable(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// reportRoutes handles request that filter lists of region e.g., /felt/report?publicID=2013p123456.
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func reportRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		reportsV1(w, r)
	default:
		notAcceptable(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// newsRoutes handles requests  for single newss e.g., /news/geonet
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func newsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1JSON:
		newsV1(w, r)
	default:
		notAcceptable(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// ok (200) - writes the content in b to the client.
func ok(w http.ResponseWriter, r *http.Request, b []byte) {
	// Haven't bothered logging sucesses.
	res.Add("2xx", 1)
	w.Write(b)
}

// notFound (404) - whatever the client was looking for we haven't got it.  The message should try
// to explain why we couldn't find that thing that they was looking for.
// Use for things that might become available e.g., a quake publicID we don't have at the moment.
func notFound(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 404")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheShort)
	w.Header().Set("Surrogate-Control", cacheShort)
	http.Error(w, message, http.StatusNotFound)
}

// notAcceptable (406) - the client requested content we don't know how to
// generate. The message should suggest content types that can be created.
func notAcceptable(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 406")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheShort)
	w.Header().Set("Surrogate-Control", cacheLong)
	http.Error(w, message, http.StatusNotAcceptable)
}

// badRequest (400) the client made a badRequest request that should not be repeated without correcting it.
// the message should explain what is badRequest about the request.
// Use for things that will never become available.
func badRequest(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 400")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheShort)
	w.Header().Set("Surrogate-Control", cacheLong)
	http.Error(w, message, http.StatusBadRequest)
}

// serviceUnavailable (500) - some sort of internal server error.
func serviceUnavailable(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 500")
	res.Add("5xx", 1)
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", http.StatusServiceUnavailable)
}
