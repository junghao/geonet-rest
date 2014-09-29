//# Region Information
//
//##/region
//
// Look up region information.  All calls return [GeoJSON](http://geojson.org/) with Polygon features.
//
//### Properties
//
// * regionID - a unique indentifier for the region.
// * title - the region title.
// * group - the region group.
//
package geojsonV1

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RegionFeatures struct {
	Features []RegionFeature
}

type RegionFeature struct {
	Properties RegionProperties
	Geometry   RegionGeometry
}

type RegionGeometry struct {
	Type string
}

type RegionProperties struct {
	RegionID, Title, Group string
}

type regionTest struct {
	regionID string
	response int
}

var rt = []regionTest{
	{"newzealand", 200},
	{"aucklandnorthland", 200},
	{"tongagrirobayofplenty", 200},
	{"gisborne", 200},
	{"hawkesbay", 200},
	{"taranaki", 200},
	{"wellington", 200},
	{"nelsonwestcoast", 200},
	{"canterbury", 200},
	{"fiordland", 200},
	{"otagosouthland", 200},
	{"ruapehu", 200},
	{"bad", 404},
}

//## Quake Regions
//
// **GET /region?type=quake**
//
// Get all quake regions.
//
//### Example request:
//
// [/region?type=quake](SERVER/region?type=quake)
//
func TestRegions(t *testing.T) {
	req, _ := http.NewRequest("GET", "/region?type=quake", nil)
	res := httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("Non 200 response code: %d", res.Code)
	}

	if res.HeaderMap.Get("Content-Type") != "application/vnd.geo+json; version=1;" {
		t.Error("incorrect Content-Type")
	}

	var f RegionFeatures

	err := json.Unmarshal([]byte(res.Body.String()), &f)
	if err != nil {
		log.Fatal(err)
	}

	if !(len(f.Features) >= 1) {
		t.Error("Found wrong number of features")
	}

	for _, feat := range f.Features {
		var g = feat.Properties.Group
		if !(g == "region" || g == "north" || g == "south") {
			t.Error("Found non quake region")
		}
	}
}

//## Single Region
//
// **GET /region/(regionID)**
//
// Get a single region.
//
//### Example request:
//
// [/region/wellington](SERVER/region/wellington)
//
func TestRegion(t *testing.T) {
	// Test a variety of routes.
	for i, rtest := range rt {
		req, _ := http.NewRequest("GET", "/region/"+rtest.regionID, nil)
		res := httptest.NewRecorder()

		serve(req, res)

		if res.Code != rtest.response {
			t.Errorf("Wrong response code for test %d: %d", i, res.Code)
		}
	}

	req, _ := http.NewRequest("GET", "/region/wellington", nil)
	res := httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("response code: %d", res.Code)
	}

	var f RegionFeatures

	err := json.Unmarshal([]byte(res.Body.String()), &f)
	if err != nil {
		log.Fatal(err)
	}

	if !(len(f.Features) == 1) {
		t.Error("Found wrong number of features")
	}

	if f.Features[0].Properties.RegionID != "wellington" {
		t.Errorf("wrong region: %s", f.Features[0].Properties.RegionID)
	}
}
