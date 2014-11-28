package main

// / request routing.  This is the app engine room.
// To add a route do the following:
// 1. Create a struct to hold the request parameters that are needed for generating content for the request (this will occasionally be empty).
// 2. Implement the query interface:
//     * Implement the validate method on the struct (this will occasionally simply return true).
//     * Implement the handle method on the struct to create and write the content.
//     * Implement the doc method.  This is used for the html docs of the api.  Put the doc struct near the query struct as code documentation.
// 3. Add the route for the request to router.
// 4. When you are ready to publish public html documentation for the route then add it's doc method to the appropriate endpoint  in api-doc.go

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	v1GeoJSON    = "application/vnd.geo+json;version=1"
	v1JSON       = "application/json;version=1"
	htmlHead     = "text/html; charset=utf-8"
	quakeLen     = 7  //  len("/quake/")
	regionLen    = 8  // len("/region/")
	endpointsLen = 19 // len("/api-docs/endpoint/")
)

var quakeRe = regexp.MustCompile(`^/quake/[0-9a-z]+$`)
var regionRe = regexp.MustCompile(`^/region/[a-z]+$`)
var feltRe = regexp.MustCompile(`^/felt/report\?publicID=[0-9a-z]+$`)
var htmlRe = regexp.MustCompile(`html`)
var endpointRe = regexp.MustCompile(`^/api-docs/endpoint/[a-z]+$`)

type query interface {
	validate(w http.ResponseWriter, r *http.Request) bool
	handle(w http.ResponseWriter, r *http.Request)
	doc() *doc
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
	switch {
	// application/vnd.geo+json;version=1
	case r.Header.Get("Accept") == v1GeoJSON:
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
	case r.Header.Get("Accept") == v1JSON:
		w.Header().Set("Content-Type", v1JSON)
		switch {
		// /news/geonet
		case r.RequestURI == "/news/geonet":
			q := &newsQuery{}
			serve(q, w, r)
		default:
			badRequest(w, r, "service not found.")
		}
	// html documentation queries.
	case htmlRe.MatchString(r.Header.Get("Accept")):
		w.Header().Set("Content-Type", htmlHead)
		w.Header().Set("Surrogate-Control", cacheMedium)
		switch {
		case r.URL.Path == "/api-docs" || r.URL.Path == "/api-docs/" || r.URL.Path == "/api-docs/index.html":
			q := &indexQuery{}
			serve(q, w, r)
		// /api-docs/endpoints/
		case endpointRe.MatchString(r.URL.Path):
			q := &endpointQuery{
				e: r.URL.Path[endpointsLen:],
			}
			serve(q, w, r)
		default:
			notFound(w, r, "page not found.")
		}
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.  Try using: "+v1GeoJSON+" or "+v1JSON)
	}
}
