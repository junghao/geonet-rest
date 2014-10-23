package main

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// reports fetches felt reports from the existing web service and prints them.
func reportsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	res, err := client.Get(feltURL + r.URL.Query().Get("publicID") + ".geojson")
	defer res.Body.Close()
	if err != nil {
		fail(w, r, err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail(w, r, err)
		return
	}

	// Felt returns a 400 when it should probably be a 404.  Tapestry quirk?
	switch {
	case 200 == res.StatusCode:
		win(w, r, b)
		return
	case 4 == res.StatusCode/100:
		nope(w, r, string(b))
		return
	case 5 == res.StatusCode/500:
		fail(w, r, errors.New("error proxying felt resports.  Shrug."))
		return
	}

	fail(w, r, errors.New("unknown response from felt."))
}
