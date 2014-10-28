package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

type gjTest struct {
	id, url, accept string
}

var gjt = []gjTest{
	{"1", "/quake/2013p407387", v1GeoJSON},
	{"2", "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good", v1GeoJSON},
	{"3", "/region/tongagrirobayofplenty", v1GeoJSON},
	{"4", "/region?type=quake", v1GeoJSON},
	{"5", "/felt/report?publicID=2013p407387", v1GeoJSON},
}

// TestGeoJSON uses geojsonlint.com to validate geoJSON
func TestGeoJSON(t *testing.T) {
	setup()
	defer teardown()

	for _, rgjt := range gjt {
		req, _ := http.NewRequest("GET", ts.URL+rgjt.url, nil)
		req.Header.Add("Accept", rgjt.accept)
		res, _ := client.Do(req)
		b, _ := ioutil.ReadAll(res.Body)

		body := bytes.NewBuffer(b)

		r, _ := client.Post("http://geojsonlint.com/validate", "application/vnd.geo+json", body)
		defer r.Body.Close()

		b, _ = ioutil.ReadAll(r.Body)

		var v Valid

		json.Unmarshal(b, &v)

		if v.Status != "ok" {
			t.Errorf("invalid geoJSON for test %s" + rgjt.id)
		}
	}
}
