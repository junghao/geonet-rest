package main

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// reports fetches felt reports from the existing web service and prints them.
func reportsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	q := &quakeQuery{
		publicID:   r.URL.Query().Get("publicID"),
		queryCount: 1,
	}

	if ok := q.validate(w, r); !ok {
		return
	}

	res, err := client.Get(feltURL + q.publicID + ".geojson")
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
