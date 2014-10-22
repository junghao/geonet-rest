package main

import (
	"net/http"
	"testing"
)

type routeTest struct {
	id, url, accept, content string
	response                 int
}

const errContent = "text/plain; charset=utf-8"

// requests without a specific Accept header ("" or "*/*") should route to the
// highest version of that part of the API.
var rt = []routeTest{
	{"1", "/quake/2013p407387", "", v1GeoJSON, 200},
	{"2", "/quake/2013p407387", "*/*", v1GeoJSON, 200},
	{"3", "/quake/2013p407387", v1GeoJSON, v1GeoJSON, 200},
	{"4", "/quake/2013p407387", "bad accept", errContent, 404},
	{"5", "/quake/2012badID", v1GeoJSON, errContent, 404},
	{"6", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", "", v1GeoJSON, 200},
	{"7", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", "*/*", v1GeoJSON, 200},
	{"8", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"9", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", "bad accept", errContent, 404},
	{"10", "/quake?regionID=newzealand&intensity=weak&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"11", "/quake?regionID=newzealand&intensity=light&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"12", "/quake?regionID=newzealand&intensity=moderate&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"13", "/quake?regionID=newzealand&intensity=strong&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"14", "/quake?regionID=newzealand&intensity=severe&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"16", "/quake?regionID=newzealand&intensity=bad&number=30&quality=best,caution,good", v1GeoJSON, errContent, 404},
	{"17", "/quake?regionID=newzealand&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"18", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"19", "/quake?regionID=newzealand&intensity=unnoticeable&number=100&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"20", "/quake?regionID=newzealand&intensity=unnoticeable&number=500&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"21", "/quake?regionID=newzealand&intensity=unnoticeable&number=1000&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"22", "/quake?regionID=newzealand&intensity=unnoticeable&number=1500&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"23", "/quake?regionID=newzealand&intensity=unnoticeable&number=999&quality=best,caution,good", v1GeoJSON, errContent, 404},
	{"24", "/quake?regionID=newzealand&intensity=unnoticeable&quality=best,caution,good", v1GeoJSON, errContent, 404},
	{"25", "/quake?regionID=newzealand&intensity=unnoticeable", v1GeoJSON, errContent, 404},
	{"26", "/quake?regionID=newzealand", v1GeoJSON, errContent, 404},
	{"27", "/quake", v1GeoJSON, errContent, 404},
	{"28", "/quake?regionID=aucklandnorthland&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"29", "/quake?regionID=tongagrirobayofplenty&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"30", "/quake?regionID=gisborne&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"31", "/quake?regionID=hawkesbay&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"32", "/quake?regionID=taranaki&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"33", "/quake?regionID=wellington&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"34", "/quake?regionID=nelsonwestcoast&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"35", "/quake?regionID=canterbury&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"36", "/quake?regionID=fiordland&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"37", "/quake?regionID=otagosouthland&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, v1GeoJSON, 200},
	{"38", "/quake?regionID=ruapehu&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, errContent, 404},
	{"39", "/quake?regionID=bad&intensity=unnoticeable&number=3&quality=best,caution,good", v1GeoJSON, errContent, 404},
	{"40", "/region/newzealand", "", v1GeoJSON, 200},
	{"41", "/region/newzealand", "*/*", v1GeoJSON, 200},
	{"42", "/region/newzealand", "bad accept", errContent, 404},
	{"43", "/region/newzealand", v1GeoJSON, v1GeoJSON, 200},
	{"44", "/region/aucklandnorthland", v1GeoJSON, v1GeoJSON, 200},
	{"45", "/region/tongagrirobayofplenty", v1GeoJSON, v1GeoJSON, 200},
	{"46", "/region/gisborne", v1GeoJSON, v1GeoJSON, 200},
	{"47", "/region/hawkesbay", v1GeoJSON, v1GeoJSON, 200},
	{"48", "/region/taranaki", v1GeoJSON, v1GeoJSON, 200},
	{"49", "/region/wellington", v1GeoJSON, v1GeoJSON, 200},
	{"50", "/region/nelsonwestcoast", v1GeoJSON, v1GeoJSON, 200},
	{"51", "/region/canterbury", v1GeoJSON, v1GeoJSON, 200},
	{"52", "/region/fiordland", v1GeoJSON, v1GeoJSON, 200},
	{"53", "/region/otagosouthland", v1GeoJSON, v1GeoJSON, 200},
	{"54", "/region/ruapehu", v1GeoJSON, v1GeoJSON, 200},
	{"55", "/region/bad", v1GeoJSON, errContent, 404},
	{"56", "/region?type=quake", "", v1GeoJSON, 200},
	{"57", "/region?type=quake", "*/*", v1GeoJSON, 200},
	{"58", "/region?type=quake", v1GeoJSON, v1GeoJSON, 200},
	{"59", "/region?type=quake", "bad accept", errContent, 404},
	{"60", "/region?type=badQuery", v1GeoJSON, errContent, 404},
	{"61", "/felt/report?publicID=2013p407387", "", v1GeoJSON, 200},
	{"62", "/felt/report?publicID=2013p407387", "*/*", v1GeoJSON, 200},
	{"63", "/felt/report?publicID=2013p407387", v1GeoJSON, v1GeoJSON, 200},
	{"64", "/felt/report?publicID=2013p407387", "bad accept", errContent, 404},
	{"65", "/felt/report?publicID=2013p999", v1GeoJSON, errContent, 404},
	{"66", "/news/geonet", "", v1JSON, 200},
	{"67", "/news/geonet", "*/*", v1JSON, 200},
	{"68", "/news/geonet", v1JSON, v1JSON, 200},
	{"69", "/news/geonet", "bad accept", errContent, 404},
}

func TestRoutes(t *testing.T) {
	setup()
	defer teardown()

	for _, r := range rt {
		req, _ := http.NewRequest("GET", ts.URL+r.url, nil)
		req.Header.Add("Accept", r.accept)
		res, _ := client.Do(req)

		if res.StatusCode != r.response {
			t.Errorf("Wrong response code for test %s: %d", r.id, res.StatusCode)
		}

		if res.Header.Get("Content-Type") != r.content {
			t.Errorf("incorrect Content-Type for test %s: %s", r.id, res.Header.Get("Content-Type"))
		}
	}
}
