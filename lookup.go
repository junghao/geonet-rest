package main

// Functions for query validation and other  lookup information.

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
)

var (
	qrV1GeoJSON []byte
	quality     map[string]int // maps for query parameter validation.  Initialized in initLookups()
	intensity   map[string]int
	number      map[string]int
	quakeRegion map[string]int
	allRegion   map[string][]byte
)

type quakeQuery struct {
	publicID   string
	queryCount int
}

type quakesQuery struct {
	regionID, intensity, number string
	quality                     []string
	queryCount                  int
}

type regionQuery struct {
	regionID   string
	queryCount int
}

// initLookups loads the query parameter validation maps.
// Some of the values are loaded from the DB.
func init() {
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

	// Connect to the DB for this init func.
	// Due to the defered close this will still need to happen again in main when the app starts.
	var err error
	db, err = sql.Open("postgres", "connect_timeout=1 user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=hazard sslmode=disable")
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()

	if err != nil {
		log.Println("Problem pinging DB - is it up and contactable.")
		log.Fatal(err)
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
	rows.Close()

	// all regions (quake and volcano)
	allRegion = make(map[string][]byte)

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
		d, err := regionV1GJ(reg)
		if err != nil {
			log.Println("Problem loading region query lookups.")
			log.Fatal(err)
		}
		allRegion[reg] = d
	}
	err = rows.Err()
	if err != nil {
		log.Println("Problem loading region query lookups.")
		log.Fatal(err)
	}
	rows.Close()

	qrV1GeoJSON, err = quakeRegionsV1GJ()
	if err != nil {
		log.Println("Problem loading quake region geojson.")
		log.Fatal(err)
	}
}

// validate checks that the quakeQuery is valid and writes errors to the responseWriter
// if not.
// the publicID is checked for slashes that may indicate extra parts in the URL.
// Returns true if the query is valid.
func (q *quakeQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// check there isn't extra stuff in the URL - like a cache buster query or extra URL parts
	if len(r.URL.Query()) != q.queryCount || strings.Contains(q.publicID, "/") {
		badRequest(w, r, "detected extra stuff in the URL.")
		return false
	}

	if q.publicID == "" {
		badRequest(w, r, "please specify a publicID")
		return false
	}

	// Check that the publicid exists in the DB.  This is needed as the geoJSON query will return empty
	// JSON for an invalid publicID.
	err := db.QueryRow("select publicid FROM qrt.quake_materialized where publicid = $1", q.publicID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid publicID: "+q.publicID)
		return false
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return false
	}
	return true
}

// validate checks if a query for quakes is valid.  Intensity can be used to validate regionIntensity
// as well (they have the same range of options) at the cost of a slightly less accurate error message.
// Returns true if the query is valid.
func (q *quakesQuery) validate(w http.ResponseWriter, r *http.Request) bool {

	// check we got the correct number of query params.  This rules out cache busters
	if len(r.URL.Query()) != 4 {
		badRequest(w, r, "detected extra stuff in the URL.")
		return false
	}

	if _, ok := number[q.number]; !ok {
		badRequest(w, r, "Invalid number: "+q.number)
		return false
	}

	if _, ok := intensity[q.intensity]; !ok {
		badRequest(w, r, "Invalid intensity: "+q.intensity)
		return false
	}

	if _, ok := quakeRegion[q.regionID]; !ok {
		badRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	for _, q := range q.quality {
		if _, ok := quality[q]; !ok {
			badRequest(w, r, "Invalid quality: "+q)
			return false
		}
	}

	return true
}

func (q *regionQuery) validate(w http.ResponseWriter, r *http.Request) bool {
	if len(r.URL.Query()) != q.queryCount || strings.Contains(q.regionID, "/") {
		badRequest(w, r, "detected extra stuff in the URL.")
		return false
	}

	// check the regionID query is valid.
	if _, ok := allRegion[q.regionID]; !ok {
		badRequest(w, r, "Invalid regionID: "+q.regionID)
		return false
	}

	return true
}
