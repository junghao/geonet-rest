package main

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// reports fetches felt reports from the existing web service and prints them.
func reportsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	// check there isn't extra stuff in the URL - like a cache buster
	if len(r.URL.Query()) != 1 {
		badRequest(w, r, "detected extra stuff in the URL.")
		return
	}

	res, err := client.Get(feltURL + r.URL.Query().Get("publicID") + ".geojson")
	defer res.Body.Close()
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	// Felt returns a 400 when it should probably be a 404.  Tapestry quirk?
	switch {
	case 200 == res.StatusCode:
		ok(w, r, b)
		return
	case 4 == res.StatusCode/100:
		notFound(w, r, string(b))
		return
	case 5 == res.StatusCode/500:
		serviceUnavailable(w, r, errors.New("error proxying felt resports.  Shrug."))
		return
	}

	serviceUnavailable(w, r, errors.New("unknown response from felt."))
}
