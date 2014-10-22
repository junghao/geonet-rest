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
	v1GeoJSON = "application/vnd.geo+json; version=1;"
	v1JSON    = "application/json; version 1;"
)

var (
	config      Config
	db          *sql.DB                     // shared DB connection pool
	client      *http.Client                // shared http client
	req         = expvar.NewInt("requests") // counters for expvar
	res         = expvar.NewMap("responses")
	quality     map[string]int // maps for query parameter validation.  Initialized in initLookups()
	intensity   map[string]int
	number      map[string]int
	quakeRegion map[string]int
	allRegion   map[string]int
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

	res.Init()
	res.Add("2xx", 0)
	res.Add("4xx", 0)
	res.Add("5xx", 0)
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", "user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=hazard sslmode=disable")
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
	client = &http.Client{
		Timeout: timeout,
	}

	initLookups()

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/quake/", quakeRoutes)
	mux.HandleFunc("/quake", quakesRoutes)
	mux.HandleFunc("/region/", regionRoutes)
	mux.HandleFunc("/region", regionsRoutes)
	mux.HandleFunc("/felt/report", reportRoutes)
	mux.HandleFunc("/news/geonet", newsRoutes)
	return get(httpgzip.NewHandler(mux))
}

// func notFound(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Cache-Control", "max-age=10")
// 	web.Nope(w, r, "service not found.")
// }

// get creates an http handler that only responds to http GET requests.  All other methods are an error.
//
//  Sets a default Cache-Control header.
func get(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Cache-Control", "max-age=10")
			h.ServeHTTP(w, r)
			return
		}
		req.Add(1)
		res.Add("4xx", 1)
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	})
}

// TODO - in these routes do we need to looks for extra stuff in the URL?
// Yes and handle errors like this http://stackoverflow.com/questions/9996767/showing-custom-404-error-page-with-standard-http-package

// quakeRoutes handles requests  for single quakes e.g., /quake/2013p12345
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func quakeRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		quakeV1(w, r)
	case "*/*":
		quakeV1(w, r)
	case "":
		quakeV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// quakesRoutes handles request that filter lists of quakes e.g., /quake?regionID=...
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func quakesRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		quakesV1(w, r)
	case "*/*":
		quakesV1(w, r)
	case "":
		quakesV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// regionRoutes handles requests  for single regions e.g., /region/newzealand
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func regionRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		regionV1(w, r)
	case "*/*":
		regionV1(w, r)
	case "":
		regionV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// regionsRoutes handles request that filter lists of region e.g., /region?type=quake.
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func regionsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		regionsV1(w, r)
	case "*/*":
		regionsV1(w, r)
	case "":
		regionsV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// reportRoutes handles request that filter lists of region e.g., /felt/report?publicID=2013p123456.
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func reportRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		reportsV1(w, r)
	case "*/*":
		reportsV1(w, r)
	case "":
		reportsV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// newsRoutes handles requests  for single newss e.g., /news/geonet
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func newsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1JSON:
		newsV1(w, r)
	case "*/*":
		newsV1(w, r)
	case "":
		newsV1(w, r)
	default:
		nope(w, r, "Can't find a route for this Accept header: "+r.Header.Get("Accept"))
	}
}

// initLookups loads the query parameter validation maps.
// Some of the values are loaded from the DB.
func initLookups() {
	quality = make(map[string]int)
	quality = map[string]int{
		"best":    1,
		"caution": 1,
		"deleted": 1,
		"good":    1,
	}

	intensity = make(map[string]int)
	intensity = map[string]int{
		"unnoticeable": 1,
		"weak":         1,
		"light":        1,
		"moderate":     1,
		"strong":       1,
		"severe":       1,
	}

	number = make(map[string]int)
	number = map[string]int{
		"3":    1,
		"30":   1,
		"100":  1,
		"500":  1,
		"1000": 1,
		"1500": 1,
	}

	// quake regions
	var reg string
	quakeRegion = make(map[string]int)

	rows, err := db.Query("select regionname FROM qrt.region where groupname in ('region', 'north', 'south')")
	if err != nil {
		log.Println("Problem loading quake region query lookups.")
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&reg)
		if err != nil {
			log.Println("Problem loading quake region query lookups.")
			log.Fatal(err)
		}
		quakeRegion[reg] = 1
	}
	err = rows.Err()
	if err != nil {
		log.Println("Problem loading quake region query lookups.")
		log.Fatal(err)
	}

	// all regions (quake and volcano)
	allRegion = make(map[string]int)

	rows, err = db.Query("select regionname FROM qrt.region")
	if err != nil {
		log.Println("Problem loading region query lookups.")
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&reg)
		if err != nil {
			log.Println("Problem loading region query lookups.")
			log.Fatal(err)
		}
		allRegion[reg] = 1
	}
	err = rows.Err()
	if err != nil {
		log.Println("Problem loading region query lookups.")
		log.Fatal(err)
	}
}

// win (200) - writes the content in b to the client.
func win(w http.ResponseWriter, r *http.Request, b []byte) {
	// Haven't bothered logging sucesses.
	res.Add("2xx", 1)
	req.Add(1)
	w.Write(b)
}

// nope (404) - whatever the client was looking for we haven't got it.  The message should try
// to explain why we couldn't find that thing that they was looking for.
func nope(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 404")
	res.Add("4xx", 1)
	req.Add(1)
	http.Error(w, message, 404)
}

// fail (500) - some sort of internal server error.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 500")
	res.Add("5xx", 1)
	req.Add(1)
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", 500)
}
