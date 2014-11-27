package main

// request routing.  This is the app engine room.
// To add a route do the following:
// 1. Create a struct to hold the request parameters (this will occasionally be empty).
// 2. Implement the query interface:
//     * Implement the validate method on the struct (this will occasionally simply return true).
//     * Implement the handle method on the struct to create and write the content.
// 3. Add the route for the request to router.

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	v1GeoJSON = "application/vnd.geo+json;version=1"
	v1JSON    = "application/json;version=1"
	quakeLen  = 7 //  len("/quake/")
	regionLen = 8 // len("/region/")
)

var quakeRe = regexp.MustCompile(`^/quake/[0-9a-z]+$`)
var regionRe = regexp.MustCompile(`^/region/[a-z]+$`)
var feltRe = regexp.MustCompile(`^/felt/report\?publicID=[0-9a-z]+$`)

type query interface {
	validate(w http.ResponseWriter, r *http.Request) bool
	handle(w http.ResponseWriter, r *http.Request)
}

func serve(q query, w http.ResponseWriter, r *http.Request) {
	if ok := q.validate(w, r); !ok {
		return
	}
	q.handle(w, r)
}

// router matches, validates, and serves http requests.
// Favour string equality over regexp when possible (performance).
// If you match on r.URL.Path you will have to check the length of r.URL.Query as well (prevent cache busters).
// If you can match r.RequestURI this will not be necessary.
// Put popular requests at the top of any switch.
func router(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	// application/vnd.geo+json;version=1
	case v1GeoJSON:
		w.Header().Set("Content-Type", v1GeoJSON)
		switch {
		// /quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good
		case r.URL.Path == "/quake" &&
			len(r.URL.Query()) == 4 &&
			r.URL.Query().Get("intensity") != "" &&
			r.URL.Query().Get("regionID") != "" &&
			r.URL.Query().Get("number") != "" &&
			r.URL.Query().Get("quality") != "":
			q := &quakesQuery{
				number:    r.URL.Query().Get("number"),
				regionID:  r.URL.Query().Get("regionID"),
				intensity: r.URL.Query().Get("intensity"),
				quality:   strings.Split(r.URL.Query().Get("quality"), ","),
			}
			serve(q, w, r)
		// /quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good
		case r.URL.Path == "/quake" && len(r.URL.Query()) == 4 &&
			r.URL.Query().Get("regionIntensity") != "" &&
			r.URL.Query().Get("regionID") != "" &&
			r.URL.Query().Get("number") != "" &&
			r.URL.Query().Get("quality") != "":
			q := &quakesRegionQuery{
				number:          r.URL.Query().Get("number"),
				regionID:        r.URL.Query().Get("regionID"),
				regionIntensity: r.URL.Query().Get("regionIntensity"),
				quality:         strings.Split(r.URL.Query().Get("quality"), ","),
			}
			serve(q, w, r)
		// /quake/2013p407387
		case quakeRe.MatchString(r.RequestURI):
			q := &quakeQuery{
				publicID: r.URL.Path[quakeLen:],
			}
			serve(q, w, r)
		// /felt/report?publicID=2013p407387
		case feltRe.MatchString(r.RequestURI):
			q := &quakeQuery{
				publicID: r.URL.Query().Get("publicID"),
			}
			serve(q, w, r)
		// /region/wellington
		case regionRe.MatchString(r.RequestURI):
			q := &regionQuery{
				regionID: r.URL.Path[regionLen:],
			}
			serve(q, w, r)
		// /region?type=quake
		case r.RequestURI == "/region?type=quake":
			q := &regionsQuery{}
			serve(q, w, r)

		default:
			badRequest(w, r, "service not found.")
		}
	// application/json;version=1
	case v1JSON:
		w.Header().Set("Content-Type", v1JSON)
		switch {
		// /news/geonet
		case r.RequestURI == "/news/geonet":
			q := &newsQuery{}
			serve(q, w, r)
		default:
			badRequest(w, r, "service not found.")
		}
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.  Try using: "+v1GeoJSON+" or "+v1JSON)
	}
}
