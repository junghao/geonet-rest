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
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
// `/quake/2013p407387`
//
func TestQuakeV1(t *testing.T) {
	setup()
	defer teardown()

	req, _ := http.NewRequest("GET", ts.URL+"/quake/2013p407387", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ := client.Do(req)
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != v1GeoJSON {
		t.Errorf("incorrect Content-Type")
	}

	var f QuakeFeatures
	err := json.Unmarshal(b, &f)
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
}

//## Quakes Possibly Felt in a Region
//
// **GET /quake?regionID=(region)&regionIntensity=(intensity)&number=(n)&quality=(quality)**
//
// Get quake information from the last 365 days.
// If no quakes are found for the query parameters then a null features array is returned.
//
//### Parameters
//
// * `regionID` - a valid quake region identifier e.g., `newzealand`.
// * `regionIntensity` - the minimum intensity in the region e.g., `weak`.  Must be one of `unnoticeable`, `weak`, `light`, `moderate`, `strong`, `severe`.
// * `number` - the maximum number of quakes to return.  Must be one of  `3`, `30`, `100`, `500`, `1000`, `1500`.
// * `quality` - a comma separated list of quality values to be included in the response; `best`, `caution`, `deleted`, `good`.
//
// *The `number` of quakes that can be returned is restricted to a range of options to improve caching.*
//
//### Example request:
//
// `/quake?regionID=newzealand&regionIntensity=weak&number=30`
//
func TestQuakesRegionV1(t *testing.T) {
	setup()
	defer teardown()

	req, _ := http.NewRequest("GET", ts.URL+"/quake?regionID=newzealand&regionIntensity=severe&number=30&quality=best,caution,good", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ := client.Do(req)
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != v1GeoJSON {
		t.Errorf("incorrect Content-Type")
	}

	var f QuakeFeatures

	err := json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f.Features) != 2 {
		t.Errorf("Found wrong number of features: %d", len(f.Features))
	}

	// Check that deleted quakes are included in the response.
	// This is a change from the existing GeoNet services.
	req, _ = http.NewRequest("GET", ts.URL+"/quake?regionID=newzealand&regionIntensity=unnoticeable&number=1000&quality=best,caution,good,deleted", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ = client.Do(req)
	defer res.Body.Close()

	b, _ = ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	err = json.Unmarshal(b, &f)
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
// *The `number` of quakes that can be returned is restricted to a range of options to improve caching.*
//
//### Example request:
//
// `/quake?regionID=newzealand&intensity=weak&number=30`
//
func TestQuakesV1(t *testing.T) {
	setup()
	defer teardown()

	// There should be 2 quakes that are felt in the Wellington region and no quakes that occur in the Wellington region.
	// This tests the difference between regionIntensity and intensity
	req, _ := http.NewRequest("GET", ts.URL+"/quake?regionID=wellington&intensity=weak&number=30&quality=best,caution,good", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ := client.Do(req)
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != v1GeoJSON {
		t.Errorf("incorrect Content-Type")
	}

	var f QuakeFeatures

	err := json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f.Features) != 0 {
		t.Errorf("Found wrong number of features: %d", len(f.Features))
	}

	req, _ = http.NewRequest("GET", ts.URL+"/quake?regionID=wellington&regionIntensity=weak&number=30&quality=best,caution,good", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ = client.Do(req)
	defer res.Body.Close()

	b, _ = ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != v1GeoJSON {
		t.Errorf("incorrect Content-Type")
	}

	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f.Features) != 2 {
		t.Errorf("Found wrong number of features: %d", len(f.Features))
	}

	// There should be 7 quakes weak and above in the Canterbury region.
	req, _ = http.NewRequest("GET", ts.URL+"/quake?regionID=canterbury&intensity=weak&number=30&quality=best,caution,good", nil)
	req.Header.Add("Accept", v1GeoJSON)
	res, _ = client.Do(req)
	defer res.Body.Close()

	b, _ = ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		t.Errorf("Non 200 error code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Type") != v1GeoJSON {
		t.Errorf("incorrect Content-Type")
	}

	err = json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f.Features) != 7 {
		t.Errorf("Found wrong number of features: %d", len(f.Features))
	}
}
