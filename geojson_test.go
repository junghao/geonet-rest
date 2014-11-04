package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

var routesGeoJSON = []routeTest{
	{
		v1GeoJSON, v1GeoJSON, cacheShort, cacheShort, http.StatusOK,
		[]route{
			{loc(), "/quake/2013p407387"},
			{loc(), "/quake?regionID=newzealand&regionIntensity=unnoticeable&number=30&quality=best,caution,good"},
			{loc(), "/quake?regionID=newzealand&intensity=unnoticeable&number=30&quality=best,caution,good"},
			{loc(), "/region/tongagrirobayofplenty"},
			{loc(), "/region?type=quake"},
			{loc(), "/felt/report?publicID=2013p407387"},
		},
	},
}

// TestGeoJSON uses geojsonlint.com to validate geoJSON
func TestGeoJSON(t *testing.T) {
	setup()
	defer teardown()

	for _, rt := range routesGeoJSON {
		for _, r := range rt.routes {
			req, _ := http.NewRequest("GET", ts.URL+r.url, nil)
			req.Header.Add("Accept", rt.accept)
			res, _ := client.Do(req)

			if res.StatusCode != rt.response {
				t.Errorf("Wrong response code for test %s: got %d expected %d", r.id, res.StatusCode, rt.response)
			}

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Problem reading body for test %s", r.id)
			}

			body := bytes.NewBuffer(b)

			res, err = client.Post("http://geojsonlint.com/validate", "application/vnd.geo+json", body)
			defer res.Body.Close()
			if err != nil {
				t.Errorf("Problem contacting geojsonlint for test %s", r.id)
			}

			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Problem reading body from geojsonlint for test %s", r.id)
			}

			var v Valid

			err = json.Unmarshal(b, &v)
			if err != nil {
				t.Errorf("Problem unmarshalling body from geojsonlint for test %s", r.id)
			}

			if v.Status != "ok" {
				t.Errorf("invalid geoJSON for test %s" + r.id)
			}
		}
	}
}
