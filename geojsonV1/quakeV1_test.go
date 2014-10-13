//# Quake Information
//
//##/quake
//
// Look up quake information.  All calls return [GeoJSON](http://geojson.org/) with Point features.
//
//### Quake Properties
//
// Each quake in the returned geojson has properties associated with it.
// Please follow this link for information about how the [quake properties](http://info.geonet.org.nz/x/J4IW) are derived.
//
// * `publicID` - the unique public identifier for this quake.
// * `time` - the origin time of the quake.
// * `depth` - the depth of the quake in km.
// * `magnitude` - the summary magnitude for the quake.  This is *not* Richter magnitude.
// * `type` - the event type; earthquake, landslide etc.
// * `agency` - the agency that located this quake.  The official GNS/GeoNet agency name for this field is WEL(*).
// * `locality` - distance and direction to the nearest locality.
// * `intensity` - the calculated [intensity](http://info.geonet.org.nz/x/b4Ih) at the surface above the quake (epicenter) e.g., `strong`.
// * `regionIntensity` - the calculated intensity at the closest locality in the region for the request.  If no region is specified for the query then this is the intensity in the `newzealand` region.
// * `quality` - the quality of this information; `best`, `good`, `caution`, `unknown`, `deleted`.
// * `modificationTime` - the modification time of this information.
//
package geojsonV1

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type QuakeFeatures struct {
	Features []QuakeFeature
}

type QuakeFeature struct {
	Properties QuakeProperties
	Geometry   QuakeGeometry
}

type QuakeGeometry struct {
	Type        string
	Coordinates []float64
}

type QuakeProperties struct {
	Publicid, Time, Modificationtime, Type, Agency string
	Locality, Intensity, Regionintensity, Quality  string
	Depth, Magnitude                               float64
}

type quakeTest struct {
	regionID, intensity, number string
	response                    int
}

var qt = []quakeTest{
	{"newzealand", "unnoticeable", "30", 200},
	{"newzealand", "weak", "30", 200},
	{"newzealand", "light", "30", 200},
	{"newzealand", "moderate", "30", 200},
	{"newzealand", "strong", "30", 200},
	{"newzealand", "severe", "30", 200},
	{"newzealand", "weak", "100", 200},
	{"newzealand", "weak", "500", 200},
	{"newzealand", "weak", "1000", 200},
	{"newzealand", "weak", "1500", 200},
	{"newzealand", "weak", "999", 404},
	{"newzealand", "bad", "30", 404},
	{"newzealand", "unnoticeable", "30", 200},
	{"aucklandnorthland", "unnoticeable", "30", 200},
	{"tongagrirobayofplenty", "unnoticeable", "30", 200},
	{"gisborne", "unnoticeable", "30", 200},
	{"hawkesbay", "unnoticeable", "30", 200},
	{"taranaki", "unnoticeable", "30", 200},
	{"wellington", "unnoticeable", "30", 200},
	{"nelsonwestcoast", "unnoticeable", "30", 200},
	{"canterbury", "unnoticeable", "30", 200},
	{"fiordland", "unnoticeable", "30", 200},
	{"otagosouthland", "unnoticeable", "30", 200},
	{"ruapehu", "unnoticeable", "30", 404},
	{"bad", "unnoticeable", "30", 404},
	{"bad", "bad", "30", 404},
	{"bad", "bad", "999", 404},
}

//## Single Quake
//
//  **GET /quake/(publicID)**
//
// Get information for a single quake.
//
//### Parameters
//
// * `publicID` - a valid quake ID e.g., `2014p715167`.
//
//### Example request:
//
// [/quake/2013p407387](SERVER/quake/2013p407387)
//
func TestQuake(t *testing.T) {
	req, _ := http.NewRequest("GET", "/quake/2013p407387", nil)
	res := httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("Non 200 error code: %d", res.Code)
	}

	if res.HeaderMap.Get("Content-Type") != "application/vnd.geo+json; version=1;" {
		t.Errorf("incorrect Content-Type")
	}

	ok, err := validateGeoJSON([]byte(res.Body.String()))
	if err != nil {
		t.Error("Problem validating GeoJSON")
	}

	if !ok {
		t.Error("Invalid GeoJSON")
	}

	var f QuakeFeatures
	err = json.Unmarshal([]byte(res.Body.String()), &f)
	if err != nil {
		log.Fatal(err)
	}

	if f.Features[0].Geometry.Type != "Point" {
		t.Error("wrong type")
	}

	if f.Features[0].Geometry.Coordinates[0] != 172.28223 {
		t.Error("wrong longitude")
	}

	if f.Features[0].Geometry.Coordinates[1] != -43.397461 {
		t.Error("wrong latitude")
	}

	if f.Features[0].Properties.Publicid != "2013p407387" {
		t.Error("incorrect publicid")
	}

	if f.Features[0].Properties.Time != "2013-05-30T15:15:37.812Z" {
		t.Error("incorrect time")
	}

	if f.Features[0].Properties.Modificationtime != "2013-06-13T23:47:04.344Z" {
		t.Error("incorrect updatetime")
	}

	if f.Features[0].Properties.Type != "earthquake" {
		t.Error("incorrect type")
	}

	if f.Features[0].Properties.Quality != "good" {
		t.Error("incorrect quality")
	}

	if f.Features[0].Properties.Intensity != "moderate" {
		t.Error("incorrect intensity")
	}

	if f.Features[0].Properties.Regionintensity != "moderate" {
		t.Error("incorrect region intensity")
	}

	if f.Features[0].Properties.Agency != "WEL(Avalon)" {
		t.Error("incorrect agency")
	}

	if f.Features[0].Properties.Locality != "15 km south-east of Oxford" {
		t.Error("incorrect locality")
	}

	if f.Features[0].Properties.Depth != 20.141276 {
		t.Error("incorrect depth")
	}

	if f.Features[0].Properties.Magnitude != 4.0252561 {
		t.Error("incorrect magnitude")
	}

	// Test an invalid quakeID.
	req, _ = http.NewRequest("GET", "/quake/2011a443", nil)
	res = httptest.NewRecorder()

	serve(req, res)

	if res.Code != 404 {
		t.Errorf("Non 404 error code: %d", res.Code)
	}
}

//## Quakes in a Region
//
// **GET /quake?regionID=(region)&intensity=(intensity)&number=(n)&quality=(quality)**
//
// Get quake information from the last 365 days.
// If no quakes are found for the query parameters then a null features array is returned.
//
//### Parameters
//
// * `regionID` - a valid quake region identifier e.g., `newzealand`.
// * `intensity` - the minimum intensity at the epicenter e.g., `weak`.  Must be one of `unnoticeable`, `weak`, `light`, `moderate`, `strong`, `severe`.
// * `number` - the maximum number of quakes to return.  Must be one of `30`, `100`, `500`, `1000`, `1500`.
// * `quality` - a comma separated list of quality values to be included in the response; `best`, `caution`, `deleted`, `good`.
//
// The `number` of quakes that can be returned is restricted to a range of options to improve caching.*
//
//### Example request:
//
// [/quake?regionID=newzealand&intensity=weak&number=30](SERVER/quake?regionID=newzealand&intensity=weak&number=30&quality=best,caution,deleted,good)
//
func TestQuakes(t *testing.T) {
	// Test a variety of routes.
	for i, qtest := range qt {
		req, _ := http.NewRequest("GET", "/quake?regionID="+qtest.regionID+"&intensity="+qtest.intensity+"&number="+qtest.number+"&quality=best,caution,good", nil)
		res := httptest.NewRecorder()

		serve(req, res)

		if res.Code != qtest.response {
			t.Errorf("Wrong response code for test %d: %d", i, res.Code)
		}
	}

	// Fetch some features and check we can decode the JSON.
	req, _ := http.NewRequest("GET", "/quake?regionID=newzealand&intensity=severe&number=30&quality=best,caution,good", nil)
	res := httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("Non 200 response code: %d", res.Code)
	}

	if res.HeaderMap.Get("Content-Type") != "application/vnd.geo+json; version=1;" {
		t.Error("incorrect Content-Type")
	}

	ok, err := validateGeoJSON([]byte(res.Body.String()))
	if err != nil {
		t.Error("Problem validating GeoJSON")
	}

	if !ok {
		t.Error("Invalid GeoJSON")
	}

	var f QuakeFeatures

	err = json.Unmarshal([]byte(res.Body.String()), &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f.Features) != 2 {
		t.Error("Found wrong number of features")
	}

	// Check that deleted quakes are included in the response.
	// This is a change from the existing GeoNet services.

	req, _ = http.NewRequest("GET", "/quake?regionID=newzealand&intensity=unnoticeable&number=1000&quality=best,caution,good,deleted", nil)
	res = httptest.NewRecorder()

	serve(req, res)

	if res.Code != 200 {
		t.Errorf("Non 200 response code: %d", res.Code)
	}

	err = json.Unmarshal([]byte(res.Body.String()), &f)
	if err != nil {
		log.Fatal(err)
	}

	var count = 0
	for _, q := range f.Features {
		if q.Properties.Quality == "deleted" {
			count++
		}
	}
	if count == 0 {
		t.Error("found no deleted quakes in the JSON.")
	}
}
