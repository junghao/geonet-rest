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
	"strconv"
	"time"
)

const (
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
	dbTime  timer
)

type Config struct {
	DataBase DataBase
	Server   Server
}

type DataBase struct {
	Host, User, Password                          string
	MaxOpenConns, MaxIdleConns, ConnectionTimeOut int
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
		if err == nil {
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

	resTime = timer{count: 0, time: 0, interval: 30 * time.Second, v: expvar.NewFloat("averageResponseTime")}
	dbTime = timer{count: 0, time: 0, interval: 30 * time.Second, v: expvar.NewFloat("averageDBResponseTime")}
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", "connect_timeout=1 user="+config.DataBase.User+
		" password="+config.DataBase.Password+
		" host="+config.DataBase.Host+
		" connect_timeout="+strconv.Itoa(config.DataBase.ConnectionTimeOut)+
		" dbname=hazard sslmode=disable")
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

	go resTime.avg()
	go dbTime.avg()

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	return get(httpgzip.NewHandler(mux))
}

// get creates an http handler that only responds to http GET requests.  All other methods are an error.
// Sets default Cache-Control and Surrogate-Control headers.
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
	log.Printf("ERROR: 500 %s", r.URL)
	res.Add("5xx", 1)
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", http.StatusServiceUnavailable)
}
