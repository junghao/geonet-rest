package main

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

type routeTest struct {
	accept, content, cache, surrogate string
	response                          int
	routes                            []route
}

type route struct {
	id, url string
}

const errContent = "text/plain; charset=utf-8"

// Any tests that expect a 200 response are run again with:
//   1. a cache buster to make sure the return a bad request (400)
//   2. extra stuff (e.g., .../bob)on the end of the URL if there is no query to make sure it returns a not acceptable (406)
//   3. the accept header set bad to make sure they return a not acceptable (406)
var routes = []routeTest{
	{
		v1GeoJSON, v1GeoJSON, cacheShort, cacheShort, http.StatusOK,
		[]route{
			{loc(), "/quake/2013p407387"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good"},
			{loc(), "/felt/report?publicID=2013p407387"},
			{loc(), "/quake?regionID=newzealand&intensity=weak&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=light&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=moderate&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=strong&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=severe&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=100&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=500&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=1000&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=1500&quality=best,caution,good"},
			{loc(), "/quake?regionID=aucklandnorthland&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=tongagrirobayofplenty&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=gisborne&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=hawkesbay&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=taranaki&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=wellington&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=nelsonwestcoast&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=canterbury&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=fiordland&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=otagosouthland&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/region/newzealand"},
			{loc(), "/region/aucklandnorthland"},
			{loc(), "/region/tongagrirobayofplenty"},
			{loc(), "/region/gisborne"},
			{loc(), "/region/hawkesbay"},
			{loc(), "/region/taranaki"},
			{loc(), "/region/wellington"},
			{loc(), "/region/nelsonwestcoast"},
			{loc(), "/region/canterbury"},
			{loc(), "/region/fiordland"},
			{loc(), "/region/otagosouthland"},
			{loc(), "/region/ruapehu"},
		},
	},
	{
		v1GeoJSON, errContent, cacheShort, cacheShort, http.StatusNotFound, // 404s that may become available
		[]route{
			{loc(), "/quake/2013p407399"},
		},
	},
	{
		v1JSON, v1JSON, cacheShort, cacheMedium, http.StatusOK,
		[]route{
			{loc(), "/news/geonet"},
		},
	},
	{
		v1GeoJSON, errContent, cacheShort, cacheLong, http.StatusBadRequest,
		[]route{
			{loc(), "/quake?regionID=newzealand&intensity=bad&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=1500&quality=best,caution,good&busta=true"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=999&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable"},
			{loc(), "/quake?regionID=newzealand"},
			{loc(), "/quake"},
			{loc(), "/quake?regionID=ruapehu&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/quake?regionID=bad&intensity=unnoticeable&number=3&quality=best,caution,good"},
			{loc(), "/region/bad"},
			{loc(), "/region?type=badQuery"},
			{loc(), "/"},
		},
	},
}

// TestRoutes tests the routes just as they are provided
func TestRoutes(t *testing.T) {
	setup()
	defer teardown()

	for _, rt := range routes {
		rt.test(t)
	}
}

// TestRoutesBuested tests the provided routes and if they should return a 200
// it adds a cache buster and tests to check they return bad request
func TestRoutesBusted(t *testing.T) {
	setup()
	defer teardown()

	var b = routeTest{"", errContent, cacheShort, cacheLong, http.StatusBadRequest,
		[]route{{"", ""}}}

	for _, rt := range routes {
		for _, r := range rt.routes {
			if rt.response == http.StatusOK {
				b.accept = rt.accept

				if strings.Contains(r.url, "?") {
					b.routes[0] = route{r.id, r.url + "&cacheBusta=1234"}
				} else {
					b.routes[0] = route{r.id, r.url + "?cacheBusta=1234"}
				}

				b.test(t)
			}
		}
	}
}

// TestRoutesExtra tests the provided routes and if they should return a 200 and they have no
// query parameters it appends extra parts on the URL and tests to check they return bad request
func TestRoutesExtra(t *testing.T) {
	setup()
	defer teardown()

	var b = routeTest{"", errContent, cacheShort, cacheLong, http.StatusBadRequest,
		[]route{{"", ""}}}

	for _, rt := range routes {
		for _, r := range rt.routes {
			if rt.response == http.StatusOK {
				if !strings.Contains(r.url, "?") {
					b.accept = rt.accept
					b.routes[0] = route{r.id, r.url + "/bob"}
					b.test(t)
				}
			}
		}
	}
}

// TestRoutesBadAccept tests all routes with an bad Accept header.
func TestRoutesBadAccept(t *testing.T) {
	setup()
	defer teardown()

	var b = routeTest{"", errContent, cacheShort, cacheLong, http.StatusNotAcceptable,
		[]route{{"", ""}}}

	for _, rt := range routes {
		for _, r := range rt.routes {
			b.routes[0] = route{r.id, r.url}
			b.test(t)
		}
	}
}

// loc returns a string representing the line that this function was called from e.g., L67
func loc() (loc string) {
	_, _, l, _ := runtime.Caller(1)
	return "L" + strconv.Itoa(l)
}

// test tests the routes in routeTest and checks response code and other header values.
func (rt routeTest) test(t *testing.T) {
	for _, r := range rt.routes {
		req, _ := http.NewRequest("GET", ts.URL+r.url, nil)
		req.Header.Add("Accept", rt.accept)
		res, _ := client.Do(req)

		if res.StatusCode != rt.response {
			t.Errorf("Wrong response code for test %s: got %d expected %d", r.id, res.StatusCode, rt.response)
		}

		if res.Header.Get("Content-Type") != rt.content {
			t.Errorf("incorrect Content-Type for test %s: %s", r.id, res.Header.Get("Content-Type"))
		}

		if res.Header.Get("Cache-Control") != rt.cache {
			t.Errorf("incorrect Cache-Control for test %s: %s", r.id, res.Header.Get("Cache-Control"))
		}

		if res.Header.Get("Surrogate-Control") != rt.surrogate {
			t.Errorf("incorrect Surrogate-Control for test %s: %s", r.id, res.Header.Get("Cache-Control"))
		}

		if !strings.Contains("Accept", res.Header.Get("Vary")) {
			t.Errorf("incorrect Vary for test %s: %s", r.id, res.Header.Get("Vary"))
		}
	}
}
